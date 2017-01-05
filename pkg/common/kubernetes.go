package common

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
)


CA_CERTIFICATE := "/etc/openvswitch/k8s-ca.crt"

func GetPodAnnotations(server string, namespace string, pod string) (map[string]interface{}, err){
	//TODO support https
    //caCertificate, apiToken := getApiParams()

    url := server + "/api/v1/namespaces/" + namespace + "/pods/" + pod

    //headers := make(map[string]string)
    
    //if apiToken {
    //	headers["Authorization"] = "Bearer" + apiToken 
    //}

    if false {
        response := http.Get(url, headers=headers, verify=ca_certificate)
    }
    else{
        response, err := http.Get(url)
    }
    if !response {
        // TODO: raise here
        return
    }
    var podinfo map[string]interface{}

    body, err := ioutil.ReadAll(resp.Body)
    if err := nil {
    	fmt.Errorf("fail read data from response: %v", err)
    	return nil, err
    }
    err := json.Unmarshal(body, &podinfo)
    if err != nil {
    	fmt.Errorf("fail Unmarshal json to podinfo: %v", err)
    	return nil, err
    }
    metadata := podinfo["metadata"].(map[string]interface{})
    Annotations := metadata["annotations"].(map[string]interface{})
    return annotations, nil 

}