package webhook

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func getSideCars() []corev1.Container {
	containers, err := GetContainerInMyConfigMap(name, namespace)
	if err != nil {
		glog.Infof("GetContainerInMyConfigMap failed in name :%s namespace: %s, use default container :%v", name, namespace, defaultContainer)
		return []corev1.Container{defaultContainer}

	}
	return containers
}

// GetContainerInMyConfigMap 从crd的 myconfigMap中获取 容器
func GetContainerInMyConfigMap(name, namespace string) ([]corev1.Container, error) {
	glog.Infof("get siecar in MyConf, name: %s, namespace :%s", name, namespace)
	clusterConf, err := rest.InClusterConfig()
	if err != nil {
		err = fmt.Errorf("get crd failed. get InClusterConfig err:%v", err)
		glog.Errorf(err.Error())
		return nil, err
	}
	dynamicClient, _ := dynamic.NewForConfig(clusterConf)
	confCrd := schema.GroupVersionResource{Group: "first.yulibaozi.com", Version: "v1beta1", Resource: "confs"}
	glog.Infof("get conf :%s", confCrd)
	unstructuredConf, err := dynamicClient.Resource(confCrd).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		err = fmt.Errorf("get unstructuredConf failed. err: %s, namespace :%s, name: %s ", err, namespace, name)
		glog.Errorf(err.Error())
		return nil, err
	}
	jsonByts, err := unstructuredConf.MarshalJSON()
	if err != nil {
		err = fmt.Errorf("unstructuredConf.MarshalJSON failed. err: %v namespace :%s, name: %s ", err, namespace, name)
		glog.Errorf(err.Error())
		return nil, err
	}
	mc := &MyConf{Data: make(map[string]string, 0)}
	err = json.Unmarshal(jsonByts, mc)
	if err != nil {
		err = fmt.Errorf("json.Unmarshal MyConf failed. err: %v namespace :%s, name: %s ", err, namespace, name)
		glog.Errorf(err.Error())
		return nil, err
	}
	if len(mc.Data) <= 0 {
		err := fmt.Errorf("data is empty in MyConf, namespace :%s, name: %s ", namespace, name)
		glog.Errorf(err.Error())
		return nil, err
	}
	containers := []corev1.Container{}
	for _, cstring := range mc.Data {
		container := corev1.Container{}
		err := json.Unmarshal([]byte(cstring), &container)
		if err != nil {
			glog.Errorf("get container failed in data map, err:%c, namespace :%s, name: %s ", err, namespace, name)
			continue
		}
		containers = append(containers, container)
	}
	if len(containers) <= 0 {
		err = fmt.Errorf("not find containers in MyConf")
		glog.Errorf(err.Error())
		return nil, err
	}
	return containers, nil
}
