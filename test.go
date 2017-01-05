package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
)

func main() {
	url := "http://10.10.101.97:8080/api/v1/namespaces/default/pods/nginx1"
	res, err := http.Get(url)
	if err != nil {
		fmt.Errorf("get err : %v", err)
	}
	body, _ := ioutil.ReadAll(res.Body)

	var podinfo map[string]interface{}
	_ = json.Unmarshal(body, &podinfo)
	metadata := podinfo["metadata"].(map[string]interface{})
	annotations := metadata["annotations"].(map[string]interface{})
	fmt.Println(annotations["ovn"])

	ip := "192.168.3.5/24"
	gateway := "192.168.3.1/24"

	_, ipnet, _ := net.ParseCIDR(ip)
	fmt.Println(ipnet.String())

	_, gatewaynet, _ := net.ParseCIDR(gateway)
	fmt.Println(gatewaynet.String())

}
