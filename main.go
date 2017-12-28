package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

import (
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

import (
	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Ec2Env struct {
	AZ           string
	InstanceID   string
	PrivateIP    string
	InstanceType string
}

func getEc2Env() *Ec2Env {
	session := session.Must(session.NewSession())
	ec2 := ec2metadata.New(session)

	if !ec2.Available() {
		return nil
	}

	idd, err := ec2.GetInstanceIdentityDocument()
	if err != nil {
		log.Printf("Can't get EC2 IIDD: %v", err)
		return nil
	}

	return &Ec2Env{
		idd.AvailabilityZone,
		idd.InstanceID,
		idd.PrivateIP,
		idd.InstanceType,
	}
}

type VirtualEnv struct {
	Virtualisation string
}

func getVirtualEnv() *VirtualEnv {
	virtwhat := exec.Command("virt-what")
	cmdOut, err := virtwhat.Output()
	if err != nil {
		/* TODO: tagged union that can represent: none, unknown, Some() */
		return &VirtualEnv{"unknown"}
	}
	/* This seems to be the simple version of ioutil.ReadAll() */
	virt := string(cmdOut)
	if virt == "" {
		return nil
	} else {
		return &VirtualEnv{virt}
	}
	/*
	* - nested virtualisation? */
}

type ContainerEnv struct {
	Containerisation string
}

func getContainerEnv() *ContainerEnv {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return &ContainerEnv{"docker"}
	} else {
		return nil
	}
	/*
	* https://stackoverflow.com/questions/20010199/determining-if-a-process-runs-inside-lxc-docker
	* - .dockerenv
	* - pid1 cgroups
	* - rkt??
	* - systemd nspawn?
	* - containerd?
	* - docker-in-docker?
	 */
}

type K8sEnv struct {
	Version string
}

func getK8sEnv() *K8sEnv {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Can't get a k8s config, probably not in a cluster? %v", err)
		return nil
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Printf("Can't create k8s clientset, despite having config: %v", err)
		return nil
	}

	css, err := clientset.CoreV1().ComponentStatuses().List(metav1.ListOptions{})
	if err != nil {
		log.Printf("Can't iterate ComponentStatuses: %v", err)
	} else {
		for _, cs := range css.Items {
			// always seems to come up empty
			log.Println(cs.String())
		}
	}

	serverVersionStruct, err := clientset.ServerVersion()
	var serverVersion string
	if err != nil {
		log.Printf("Got k8s server connection, but can't get version info: %v", err)
		serverVersion = "<unknown>"
	} else {
		serverVersion = serverVersionStruct.String()
	}

	return &K8sEnv{
		serverVersion,
	}
}

/* Get the IP of the interface through which the default route goes */
func getDefaultIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		log.Println(err)
		return "<unknown>"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func getHostname() (hostname string) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Can't get hostname: %v", err)
		hostname = "<unknown>"
	}
	return
}

type PageData struct {
	HostIp    string
	Hostname  string
	HostOs    string
	HostArch  string
	ClientIp  string
	Container *ContainerEnv
	K8s       *K8sEnv
	Virtual   *VirtualEnv
	Aws       *Ec2Env
}

func getSiblingPath(file string) string {
	me, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(filepath.Dir(me), file)
}

func handler(w http.ResponseWriter, r *http.Request) {
	hostip := getDefaultIp()
	hostname := getHostname()
	clientip := r.RemoteAddr

	data := &PageData{
		Hostname:  hostname,
		HostIp:    hostip,
		HostOs:    runtime.GOOS,
		HostArch:  runtime.GOARCH,
		ClientIp:  clientip,
		Container: getContainerEnv(),
		K8s:       getK8sEnv(),
		Virtual:   getVirtualEnv(),
		Aws:       getEc2Env(),
	}

	var t *template.Template
	var err error
	t, err = template.ParseFiles(getSiblingPath("main.html"))
	if err != nil {
		t, err = template.ParseFiles("main.html")
		if err != nil {
			log.Fatalf("%v. Hint: ensure main.html is alongside executable or in $PWD.", err)
		}
	}
	t.Execute(w, data)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Serving...")
	http.ListenAndServe(":8080", nil)
}
