package main

import (
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker" // Updated to v4
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		return createResources(ctx)
	})
}

func createResources(ctx *pulumi.Context) error {
	// 1. PersistentVolumeClaim
	pvc, err := corev1.NewPersistentVolumeClaim(ctx, "sqlite-pvc", &corev1.PersistentVolumeClaimArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("sqlite-pvc"),
		},
		Spec: &corev1.PersistentVolumeClaimSpecArgs{
			AccessModes: pulumi.StringArray{pulumi.String("ReadWriteOnce")},
			Resources: &corev1.VolumeResourceRequirementsArgs{
				Requests: pulumi.StringMap{
					"storage": pulumi.String("1Gi"),
				},
			},
			StorageClassName: pulumi.String("microk8s-hostpath"),
		},
	})
	if err != nil {
		return err
	}
	// Inside createResources...
	cwd, _ := os.Getwd()
	// This creates a solid path like /home/ihor/GolandProjects/pulumiEval/counter
	absCounterDir := filepath.Join(cwd, "counter")
	absDockerfile := filepath.Join(absCounterDir, "Dockerfile")

	counterImage, err := docker.NewImage(ctx, "counter-image", &docker.ImageArgs{
		Build: &docker.DockerBuildArgs{
			Context:    pulumi.String(absCounterDir),
			Dockerfile: pulumi.String(absDockerfile),
			Platform:   pulumi.String("linux/amd64"),
			// FORCE BUILDER V1 TO BYPASS NETWORK ERRORS
			BuilderVersion: docker.BuilderVersionBuilderV1,
		},
		ImageName: pulumi.String("localhost:32000/counter-server:latest"),
		Registry: &docker.RegistryArgs{
			Server: pulumi.String("localhost:32000"),
		},
	})
	if err != nil {
		return err
	}

	// 3. Deployment
	_, err = appsv1.NewDeployment(ctx, "counter-deployment", &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("counter-deployment"),
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("counter"),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("counter"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String("counter"),
							Image: counterImage.ImageName, // Now this variable exists!
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{ContainerPort: pulumi.Int(8080)},
							},
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("sqlite-storage"),
									MountPath: pulumi.String("/data"),
								},
							},
						},
					},
					Volumes: corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("sqlite-storage"),
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSourceArgs{
								ClaimName: pvc.Metadata.Name().Elem(),
							},
						},
					},
					Affinity: &corev1.AffinityArgs{
						NodeAffinity: &corev1.NodeAffinityArgs{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelectorArgs{
								NodeSelectorTerms: corev1.NodeSelectorTermArray{
									&corev1.NodeSelectorTermArgs{
										MatchExpressions: corev1.NodeSelectorRequirementArray{
											&corev1.NodeSelectorRequirementArgs{
												Key:      pulumi.String("node-role.kubernetes.io/worker"),
												Operator: pulumi.String("In"),
												Values:   pulumi.StringArray{pulumi.String("true")},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// 4. Service
	svc, err := corev1.NewService(ctx, "counter-service", &corev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("counter-service"),
		},
		Spec: &corev1.ServiceSpecArgs{
			Type: pulumi.String("NodePort"),
			Selector: pulumi.StringMap{
				"app": pulumi.String("counter"),
			},
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(8080), // Matched to the Go server port
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// Export the allocated NodePort
	ctx.Export("nodePort", svc.Spec.Ports().Index(pulumi.Int(0)).NodePort().ApplyT(func(p *int) int {
		if p == nil {
			return 0
		}
		return *p
	}).(pulumi.IntOutput))

	// Export the Image Name
	ctx.Export("deployedImage", counterImage.ImageName)

	return nil
}
