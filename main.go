package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"net/http"
	_ "net/http/pprof"

	gadgetcontext "github.com/inspektor-gadget/inspektor-gadget/pkg/gadget-context"
	_ "github.com/inspektor-gadget/inspektor-gadget/pkg/operators/ebpf"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/operators/localmanager"
	ocihandler "github.com/inspektor-gadget/inspektor-gadget/pkg/operators/oci-handler"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/operators/socketenricher"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/runtime/local"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/utils/host"
	"github.com/sirupsen/logrus"
	"oras.land/oras-go/v2/content/oci"
)

func do() error {
	if err := host.Init(host.Config{AutoMountFilesystems: true}); err != nil {
		return fmt.Errorf("failed to init host: %w", err)
	}

	// const opPriority = 50000
	// myOperator := simple.New("myHandler",
	// 	simple.WithPriority(opPriority),
	// 	simple.OnInit(func(gadgetCtx operators.GadgetContext) error {
	// 		for _, d := range gadgetCtx.GetDataSources() {
	// 			jsonFormatter, _ := igjson.New(d,
	// 				igjson.WithShowAll(true),
	// 			)

	// 			d.Subscribe(func(source datasource.DataSource, data datasource.Data) error {
	// 				jsonOutput := jsonFormatter.Marshal(data)
	// 				fmt.Println(len(jsonOutput))
	// 				return nil
	// 			}, opPriority)
	// 		}
	// 		return nil
	// 	}))

	// Create the local manager operator
	localManagerOp := localmanager.LocalManagerOperator
	localManagerParams := localManagerOp.GlobalParamDescs().ToParams()
	// Needed for owner and pod's label enrichment.
	if err := localManagerParams.Get(localmanager.EnrichWithK8sApiserver).Set("true"); err != nil {
		return err
	}
	if err := localManagerOp.Init(localManagerParams); err != nil {
		return fmt.Errorf("init local manager: %w", err)
	}
	defer localManagerOp.Close()

	ociStore, err := oci.NewFromTar(context.Background(), "trace_dns.tar")
	if err != nil {
		return err
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	gadgetCtx := gadgetcontext.New(
		context.Background(),
		"ghcr.io/inspektor-gadget/gadget/trace_dns:v0.38.0",
		gadgetcontext.WithOrasReadonlyTarget(ociStore),
		gadgetcontext.WithLogger(logger),
		gadgetcontext.WithDataOperators(
			ocihandler.OciHandler,
			localManagerOp,
			&socketenricher.SocketEnricher{},
			// myOperator,
		),
	)

	runtime := local.New()
	if err := runtime.Init(nil); err != nil {
		return fmt.Errorf("runtime init: %w", err)
	}
	defer runtime.Close()

	params := map[string]string{
		"operator.oci.ebpf.paths":        "true",
		"operator.oci.ebpf.tracer-group": "dns",
	}
	if err := runtime.RunGadget(gadgetCtx, nil, params); err != nil {
		return fmt.Errorf("running gadget: %w", err)
	}

	return nil
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Your application code here
	if err := do(); err != nil {
		fmt.Printf("Error running application: %s\n", err)
		os.Exit(1)
	}
}
