package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

const version = "1.0.0"

func main() {
	var namespace string
	var outputFile string

	rootCmd := &cobra.Command{
		Use:   "k8s-dump",
		Short: "Dump Kubernetes resources to a YAML file",
		Run: func(cmd *cobra.Command, args []string) {
			// Get Kubernetes config
			config, err := rest.InClusterConfig()
			if err != nil {
				home := os.Getenv("HOME")
				kubeconfig := filepath.Join(home, ".kube", "config")
				config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
				if err != nil {
					panic(err.Error())
				}
			}

			// Create Kubernetes client
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				panic(err.Error())
			}

			// List all resource types
			apiResources, err := clientset.Discovery().ServerPreferredNamespacedResources()
			if err != nil {
				panic(err.Error())
			}

			// Open output file
			file, err := os.Create(outputFile)
			if err != nil {
				panic(err.Error())
			}
			defer file.Close()

			// Write each resource type to the output file
			for _, apiResourceList := range apiResources {
				for _, apiResource := range apiResourceList.APIResources {
					gvk := apiResourceList.GroupVersion + "/" + apiResource.Kind
					fmt.Println("Dumping", gvk, "to", outputFile)
					// Get the resource list
					resourceList, err := clientset.RESTClient().Get().
						AbsPath("/apis", apiResourceList.GroupVersion, "namespaces", namespace, apiResource.Name).
						DoRaw(context.TODO())
					if err != nil {
						fmt.Println("Error getting resource:", err)
						continue
					}

					// Write resource to file
					file.WriteString(fmt.Sprintf("# Resource: %s\n", gvk))
					yamlBytes, err := yaml.JSONToYAML(resourceList)
					if err != nil {
						fmt.Println("Error converting to YAML:", err)
						continue
					}
					file.Write(yamlBytes)
					file.WriteString("---\n")
				}
			}

			fmt.Printf("All resources in namespace %s have been dumped to %s\n", namespace, outputFile)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "The namespace to dump resources from")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "outputFile", "o", "all-resources.yaml", "The output file to write the resources to")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of the program",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Version:", version)
		},
	}

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
