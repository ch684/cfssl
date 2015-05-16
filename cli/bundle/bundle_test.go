package bundle

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

func bundler_test() {
	gopath := os.Getenv("GOPATH")
	cfssl := path.Join(gopath, "bin", "cfssl")
	testdata := path.Join(gopath, "src", "github.com", "cloudflare", "cfssl", "bundler", "testdata")
	pkg := path.Join(gopath, "src", "github.com", "cfssl_trust")
	//cfssl bundle [-ca-bundle file] [-int-bundle file] [-key keyfile] [-flavor int] [-metadata file] CERT
	out1, err1 := exec.Command(cfssl, "bundle", 
		"-cert="+path.Join(testdata, "bunnings.pem"), 
		"-ca-bundle="+path.Join(pkg, "ca-bundle.crt"), 
		"-int-bundle="+path.Join(pkg, "int-bundle.crt"), 
		"-metadata="+path.Join(pkg, "ca-bundle.crt.metadata").CombinedOutput()
	fmt.Printf("cfssl bundle [-ca-bundle file] [-int-bundle file] [-metadata file] [-key keyfile] [-flavor int]  CERT\n%s\n", string(out1))
	fmt.Println("Error: ", err1)

	//cfssl bundle -domain domain_name [-ip ip_address] [-ca-bundle file] [-int-bundle file] [-metadata file]
	out2, err2 := exec.Command(cfssl, "bundle", 
		"-domain=www.google.com", 
		"-ca-bundle="+path.Join(pkg, "ca-bundle.crt"), 
		"-int-bundle="+path.Join(pkg, "int-bundle.crt"_, 
		"-metadata="+path.Join(pkg, "ca-bundle.crt.metadata").CombinedOutput()
	fmt.Printf("cfssl bundle -domain domain_name [-ip ip_address] [-ca-bundle file] [-int-bundle file] [-metadata file]\n%s\n", string(out2))
	fmt.Println("Error: ", err2)