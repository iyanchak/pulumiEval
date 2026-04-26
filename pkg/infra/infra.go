package infra

import (
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateResources(ctx *pulumi.Context) error {
	// 1. PersistentVolumeClaim
	pvc, err := corev1.NewPersistentVolumeClaim(ctx, "sqlite-pvc", &corev1.PersistentVolumeClaimArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("sqlite-pvc"),
			Annotations: pulumi.StringMap{
				"pulumi.com/skipAwait": pulumi.String("true"),
			},
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

	cwd, _ := os.Getwd()
	absCounterDir := filepath.Join(cwd, "cmd", "counter")

	imageName := "counter-server:latest"

	// 2. Build local Docker Image
	buildCmd, err := local.NewCommand(ctx, "build-counter", &local.CommandArgs{
		Create: pulumi.Sprintf("pack build %s --builder gcr.io/buildpacks/builder:google-22 --buildpack google.go.runtime --buildpack google.go.build --env CGO_ENABLED=0 --path %s", pulumi.String(imageName), pulumi.String(absCounterDir)),
	})
	if err != nil {
		return err
	}

	// Side-load the image into microk8s
	importImage, err := local.NewCommand(ctx, "importImage", &local.CommandArgs{
		Create: pulumi.Sprintf("bash -c 'docker save %s | sudo microk8s ctr images import -'", pulumi.String(imageName)),
	}, pulumi.DependsOn([]pulumi.Resource{buildCmd}))
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
							Name:            pulumi.String("counter"),
							Image:           pulumi.String(imageName),
							ImagePullPolicy: pulumi.String("Never"), // Use local image for microk8s
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
	}, pulumi.DependsOn([]pulumi.Resource{importImage}))
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
					TargetPort: pulumi.Int(8080),
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
	ctx.Export("deployedImage", pulumi.String(imageName))

	return nil
}
