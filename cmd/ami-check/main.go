package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	nodecheck "github.com/kuberhealthy/kuberhealthy/v3/pkg/nodecheck"
	log "github.com/sirupsen/logrus"
)

// main wires configuration, dependencies, and executes the AMI check.
func main() {
	// Parse configuration from environment variables.
	cfg, err := parseConfig()
	if err != nil {
		reportFailure([]string{err.Error()})
		return
	}

	// Create a context bounded by the check deadline.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckTimeLimit)
	defer cancel()

	// Start handling OS signals for graceful exits.
	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	go handleSignals(signalChan)

	// Wait for the Kuberhealthy endpoint to be reachable.
	err = nodecheck.WaitForKuberhealthy(ctx)
	if err != nil {
		log.Errorln("Error waiting for kuberhealthy endpoint to be contactable by checker pod with error:", err.Error())
	}

	// Build the AWS session for the check.
	awsSession, err := createAWSSession()
	if err != nil {
		reportFailure([]string{err.Error()})
		return
	}

	// Catch unexpected panics and report them.
	defer recoverAndReport()

	// Run the main AMI check logic.
	err = runCheck(cfg, awsSession)
	if err != nil {
		reportFailure([]string{err.Error()})
		return
	}

	// Report success if no errors were found.
	reportSuccess()
}

// handleSignals handles termination signals for the checker pod.
func handleSignals(signalChan chan os.Signal) {
	// Wait for the first signal and exit immediately.
	sig := <-signalChan
	log.Infoln("Received an interrupt signal from the signal channel.")
	log.Debugln("Signal received was:", sig.String())
	log.Infoln("Shutting down.")
	os.Exit(0)
}

// recoverAndReport captures panics and reports them to Kuberhealthy.
func recoverAndReport() {
	// Read the panic value if one occurred.
	recovered := recover()
	if recovered == nil {
		return
	}

	log.Infoln("Recovered panic:", recovered)
	message := "panic: " + stringify(recovered)
	reportFailure([]string{message})
}

// reportFailure reports failed check results to Kuberhealthy.
func reportFailure(errors []string) {
	// Log and report failure errors.
	log.Errorln("Reporting errors to Kuberhealthy:", errors)
	err := checkclient.ReportFailure(errors)
	if err != nil {
		log.Fatalln("error reporting to kuberhealthy:", err.Error())
	}
}

// reportSuccess reports successful check results to Kuberhealthy.
func reportSuccess() {
	// Log and report success.
	log.Infoln("Reporting success to Kuberhealthy.")
	err := checkclient.ReportSuccess()
	if err != nil {
		log.Fatalln("error reporting to kuberhealthy:", err.Error())
	}
}

// stringify converts panic values to strings for logging.
func stringify(value interface{}) string {
	// Handle string values directly.
	text, ok := value.(string)
	if ok {
		return text
	}

	// Handle error values explicitly.
	err, ok := value.(error)
	if ok {
		return err.Error()
	}

	// Fallback to formatted output.
	return "unknown panic"
}
