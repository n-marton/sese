package main

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Secrets struct {
	Secrets []SecretToSync
}

type SecretToSync struct {
	Sourcesecret     string
	Sourcenamespace  string
	Targetprefix     string
	Targetnamespaces []string
}

func (c *Secrets) getConf() *Secrets {

	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("configFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(configFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func inArray(array []string, element string) bool {
	for _, arrayelement := range array {
		if element == arrayelement {
			return true
		}
	}
	return false
}

func getNamespaces() []string {
	var nsList []string
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error getting k8s cluster config: " + err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error setting k8s cluster client: " + err.Error())
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error getting namespaces: " + err.Error())
	}
	for _, ns := range namespaces.Items {
		nsList = append(nsList, ns.Name)
	}
	return nsList
}

func getSecret(namespace string, name string) *v1.Secret {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error getting k8s cluster config: " + err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error setting k8s cluster client: " + err.Error())
	}
	log.Printf("Getting secret: " + namespace + "/" + name)
	secret, err := clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error getting secret at: " + namespace + "/" + name + "\n Error: " + err.Error())
	}
	return secret
}

func checkSecretExistance(namespace string, name string) bool {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error getting k8s cluster config: " + err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error setting k8s cluster client: " + err.Error())
	}
	log.Printf("Checking secret existance: " + namespace + "/" + name)
	_, err = clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Secret: " + namespace + "/" + name + " doesn't exist")
		return false
	}
	log.Printf("Secret: " + namespace + "/" + name + " exist")
	return true

}

func createSecret(namespace string, secret *v1.Secret, prefix string) {
	log.Printf("Trying to create: " + namespace + "/" + secret.Name)
	var rawSecret v1.Secret
	rawSecret.Namespace = namespace
	rawSecret.Name = prefix + secret.Name
	rawSecret.Data = secret.Data
	rawSecret.Type = secret.Type
	rawSecret.TypeMeta = secret.TypeMeta
	rawSecret.StringData = secret.StringData
	rawSecret.Labels = secret.Labels
	rawSecret.Kind = secret.Kind
	rawSecret.APIVersion = secret.APIVersion
	rawSecret.Annotations = secret.Annotations
	var newSecret *v1.Secret = &rawSecret
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error getting k8s cluster config: " + err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error setting k8s cluster client: " + err.Error())
	}
	_, err = clientset.CoreV1().Secrets(namespace).Create(newSecret)
	if err != nil {
		log.Fatalf("There is an error while creating: " + namespace + "/" + prefix + secret.Name + "\n Error: " + err.Error())
	} else {
		log.Printf("Creating of : " + namespace + "/" + prefix + secret.Name + "has finished")
	}
}

func updateSecret(namespace string, secret *v1.Secret, prefix string) {
	log.Printf("Trying to update: " + namespace + "/" + secret.Name)
	var rawSecret v1.Secret
	rawSecret.Namespace = namespace
	rawSecret.Name = prefix + secret.Name
	rawSecret.Data = secret.Data
	rawSecret.Type = secret.Type
	rawSecret.TypeMeta = secret.TypeMeta
	rawSecret.StringData = secret.StringData
	rawSecret.Labels = secret.Labels
	rawSecret.Kind = secret.Kind
	rawSecret.APIVersion = secret.APIVersion
	rawSecret.Annotations = secret.Annotations
	var newSecret *v1.Secret = &rawSecret
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error getting k8s cluster config: " + err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error setting k8s cluster client: " + err.Error())
	}
	_, err = clientset.CoreV1().Secrets(namespace).Update(newSecret)
	if err != nil {
		log.Fatalf("There is an error while updating: " + namespace + "/" + prefix + secret.Name + "\n Error: " + err.Error())
	} else {
		log.Printf("Updating of : " + namespace + "/" + prefix + secret.Name + "has finished")
	}
}

func handleSecrets() {
	var c Secrets
	log.Printf("Getting configuration")
	c.getConf()
	log.Printf("Getting namespaces")
	nsList := getNamespaces()
	log.Printf("Found namespaces: " + strings.Join(nsList, ", "))
	for _, sync := range c.Secrets {
		checksourceexistance := checkSecretExistance(sync.Sourcenamespace, sync.Sourcesecret)
		if checksourceexistance {
			sourceSecret := getSecret(sync.Sourcenamespace, sync.Sourcesecret)
			if sync.Targetnamespaces == nil {
				for _, target := range nsList {
					if sync.Sourcenamespace == target {
						log.Printf("Skipping source namespace: " + sync.Sourcenamespace)
					} else {
						checktargetexistance := checkSecretExistance(target, sync.Targetprefix+sync.Sourcesecret)
						if !checktargetexistance {
							createSecret(target, sourceSecret, sync.Targetprefix)
						} else {
							targetSecretCheck := getSecret(target, sync.Targetprefix+sync.Sourcesecret)
							if !reflect.DeepEqual(sourceSecret.Data, targetSecretCheck.Data) || !reflect.DeepEqual(sourceSecret.Labels, targetSecretCheck.Labels) || !reflect.DeepEqual(sourceSecret.Annotations, targetSecretCheck.Annotations) {
								updateSecret(target, sourceSecret, sync.Targetprefix)
							} else {
								log.Printf("Sync is fine for: " + target + "/" + sync.Targetprefix + sync.Sourcesecret)
							}
						}
					}
				}
			} else {
				for _, target := range sync.Targetnamespaces {
					checktargetexistance := checkSecretExistance(target, sync.Targetprefix+sync.Sourcesecret)
					if !checktargetexistance {
						inarray := inArray(nsList, target)
						if inarray {
							createSecret(target, sourceSecret, sync.Targetprefix)
						} else {
							log.Printf("Skipping namespace: " + target + " it doesn't exist")
						}

					} else {
						targetSecretCheck := getSecret(target, sync.Targetprefix+sync.Sourcesecret)
						if !reflect.DeepEqual(sourceSecret.Data, targetSecretCheck.Data) || !reflect.DeepEqual(sourceSecret.Labels, targetSecretCheck.Labels) || !reflect.DeepEqual(sourceSecret.Annotations, targetSecretCheck.Annotations) {
							updateSecret(target, sourceSecret, sync.Targetprefix)
						} else {
							log.Printf("Sync is fine for: " + target + "/" + sync.Sourcesecret)
						}
					}
				}
			}
		} else {
			log.Printf("Source secret at: " + sync.Sourcenamespace + "/" + sync.Sourcesecret + " doesn't exists")
		}
	}
}

func main() {
	cycle := os.Getenv("CYCLE")
	if cycle == "" {
		cycle = "300"
	}
	cycleint, err := strconv.Atoi(cycle)
	if err != nil {
		log.Fatalf("Error setting cycle time: " + err.Error())
	} else {
		log.Printf("Cycle time set to: " + cycle + " seconds")
	}
	for {

		handleSecrets()
		time.Sleep(time.Duration(int32(cycleint)) * time.Second)
	}
}
