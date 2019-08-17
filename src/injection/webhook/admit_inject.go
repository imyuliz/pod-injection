package webhook

import (
	"encoding/json"
	"net/http"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServerPodInjection server
func ServerPodInjection(w http.ResponseWriter, r *http.Request) {
	glog.Infof("input request")
	serve(w, r, injection)
}

// PodInjection 注入逻辑
func injection(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	glog.Infoln("mutating pods")
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource {
		glog.Errorf(" expect resource to be %v, but request resource is :%v", podResource, ar.Request.Resource)
		return nil
	}
	glog.Infof("resource is pod")
	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		glog.Errorln(err)
		return toAdmissionResponse(err)
	}

	// 判断是否需要注入
	if needInject(pod) {
		patchBytes, err := injectOperate(pod)
		if err != nil {
			glog.Errorf("injectOperate failed. err:%v,pod:%s, namespace:%s", err, pod.Name, pod.Namespace)
			return toAdmissionResponse(err)
		}
		glog.Infof("pod: %s, namespace: %sinjectOperate patch:%s", pod.Name, pod.Namespace, string(patchBytes))
		pt := v1beta1.PatchTypeJSONPatch
		return &v1beta1.AdmissionResponse{Allowed: true, Patch: patchBytes, PatchType: &pt}
	}
	glog.Infof("not need inject, skipping. pod:%s, namespace:%s ", pod.Name, pod.Namespace)
	return &v1beta1.AdmissionResponse{Allowed: true}
}

func needInject(pod corev1.Pod) bool {
	annots := pod.GetAnnotations()
	if len(annots) <= 0 {
		return true
	}
	need := annots[INJECTEDKEY] == injectedAnnotations[INJECTEDKEY] || annots[NOINJECTEDKEY] == injectedAnnotations[NOINJECTEDKEY]
	return !need
}

func injectOperate(pod corev1.Pod) ([]byte, error) {
	glog.Infof("注入sidecar和annotations in pod: %s, namespace:%s...", pod.Name, pod.Namespace)
	var containers = []corev1.Container{}
	if len(pod.Spec.Containers) > 0 {
		for i := range pod.Spec.Containers {
			containers = append(containers, pod.Spec.Containers[i])
		}
	}
	containers = append(containers, getSideCars()...)
	containerPatch := patchOperation{Op: OpAdd, Path: PathContainers, Value: containers}
	// 注入容器
	patchs := []patchOperation{containerPatch}
	//注入annotations
	annotationsPatch := patchOperation{
		Op:    OpAdd,
		Path:  PathAnnotations,
		Value: map[string]string{INJECTEDKEY: injectedAnnotations[INJECTEDKEY]},
	}
	patchs = append(patchs, annotationsPatch)
	return json.Marshal(patchs)
}

func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
}
