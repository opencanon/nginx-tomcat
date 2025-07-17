package operator

import (
	"context"
	"fmt"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	mytomcat "opencanon.com/api/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// Define the spew configuration
	scs3 = spew.ConfigState{
		Indent:                  "  ",
		DisableMethods:          true,
		DisablePointerMethods:   true,
		DisablePointerAddresses: true,
		MaxDepth:                9,
	}

	// Define the zap encoder configuration
	encoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
)

func reconcileBinding(ctx context.Context, client client.Client, tomcat *mytomcat.Tomcat) error {

	log := logf.FromContext(ctx)
	log.Info("==reconcileBinding==")

	log.Info("Post-reconcile", "namespace", tomcat.GetNamespace(), "name", tomcat.GetName())

	myDeployment := &appsv1.Deployment{}
	deployName := fmt.Sprintf("%s-tomcat", tomcat.Name)
	if err := client.Get(ctx, types.NamespacedName{Namespace: tomcat.Namespace, Name: deployName}, myDeployment); err != nil {
		return err
	}
	n0 := myDeployment.Spec.Replicas
	n1 := int32(tomcat.Spec.Replicas)
	if *n0 != n1 {
		*myDeployment.Spec.Replicas = n1
		if err := client.Update(ctx, myDeployment); err != nil {
			return err
		}
	}
	return nil

}

func resourcesRequesetBinding(ctx context.Context, client client.Client, tomcat *mytomcat.Tomcat) error {
	log := logf.FromContext(ctx)
	log.Info("==resourcesRequesetBinding==")

	myDeployment := &appsv1.Deployment{}

	deployName := fmt.Sprintf("%s-tomcat", tomcat.Name)
	if err := client.Get(ctx, types.NamespacedName{Namespace: tomcat.Namespace, Name: deployName}, myDeployment); err != nil {
		return err
	}
	c0, m0, _ := getMemory(myDeployment)
	m1 := tomcat.Spec.Resources.Requests.Memory

	c1 := tomcat.Spec.Resources.Requests.Cpu

	if !(m0.String() == m1 && c0.String() == c1) {
		setMemCpuRequest(myDeployment, c1, m1)
		if err := client.Update(ctx, myDeployment); err != nil {
			return err
		}
	}
	return nil

}

func getMemory(deployment *appsv1.Deployment) (resource.Quantity, resource.Quantity, error) {
	// Check if Containers slice is non-empty
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return resource.Quantity{}, resource.Quantity{}, fmt.Errorf("no containers found in deployment")
	}

	// Get the first container's Resources
	resources := deployment.Spec.Template.Spec.Containers[0].Resources

	// Check if Requests map is initialized and contains Memory
	if resources.Requests == nil {
		return resource.Quantity{}, resource.Quantity{}, fmt.Errorf("no resource requests defined")
	}

	memory, exists := resources.Requests[corev1.ResourceMemory]
	if !exists {
		return resource.Quantity{}, resource.Quantity{}, fmt.Errorf("memory request not defined")
	}

	cpu, exists := resources.Requests[corev1.ResourceCPU]
	if !exists {
		return resource.Quantity{}, resource.Quantity{}, fmt.Errorf("memory request not defined")
	}

	return cpu, memory, nil
}

func setMemCpuRequest(deployment *appsv1.Deployment, c0 string, m0 string) {
	// Ensure first container exists
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		deployment.Spec.Template.Spec.Containers = []corev1.Container{{}}
	}

	container := deployment.Spec.Template.Spec.Containers[0]

	// Set correct resource names
	container.Resources.Requests[corev1.ResourceMemory] = resource.MustParse(m0)
	container.Resources.Requests[corev1.ResourceCPU] = resource.MustParse(c0)

}

func myBinding2(ctx context.Context, client client.Client, obj *mytomcat.Tomcat) error {
	log := logf.FromContext(ctx)
	log.Info("==6688 22==")

	// Create a file-based logger
	logger, err := createFileLogger("spew_dump_Post_Rec.log", zapcore.InfoLevel, encoderConfig)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	defer logger.Sync()

	// Dump the object to the file-based logger
	dumpToLogger(logger, &scs3, ctx)
	dumpToLogger(logger, &scs3, obj)

	// Simulate some work
	time.Sleep(1 * time.Second)

	return nil
}

// createFileLogger creates a zap.Logger that writes to the specified file.
func createFileLogger(filename string, level zapcore.Level, encoderConfig zapcore.EncoderConfig) (*zap.Logger, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	fileSyncer := zapcore.AddSync(file)
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		fileSyncer,
		level,
	)
	return zap.New(core), nil
}

// dumpToLogger formats an object using spew and logs it to a zap.Logger.
func dumpToLogger(logger *zap.Logger, scs *spew.ConfigState, obj interface{}) {
	// Use Sdump to get the formatted string
	formatted := scs.Sdump(obj)
	// Log the formatted string at Info level (or another level as needed)
	logger.Info("Spew dump", zap.String("dump", formatted))
}
