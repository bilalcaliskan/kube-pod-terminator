package cmd

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bilalcaliskan/kube-pod-terminator/internal/version"

	"github.com/bilalcaliskan/kube-pod-terminator/internal/k8s"
	"github.com/bilalcaliskan/kube-pod-terminator/internal/logging"
	"github.com/bilalcaliskan/kube-pod-terminator/internal/options"
	"github.com/dimiro1/banner"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	logger            *zap.Logger
	kubeConfigPathArr []string
	opts              *options.KubePodTerminatorOptions
	ver               = version.Get()
)

func init() {
	opts = options.GetKubePodTerminatorOptions()

	rootCmd.Flags().BoolVarP(&opts.InCluster, "inCluster", "", false, "specify if kube-pod-terminator is running in cluster")
	rootCmd.Flags().StringVarP(&opts.KubeConfigPaths, "kubeConfigPaths", "", filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		"comma separated list of kubeconfig file paths to access with the cluster")
	rootCmd.Flags().StringVarP(&opts.Namespace, "namespace", "", "all", "target namespace to run on")
	rootCmd.Flags().Int32VarP(&opts.TickerIntervalMinutes, "tickerIntervalMinutes", "", 5, "interval of scheduled job to run")
	rootCmd.Flags().IntVarP(&opts.ChannelCapacity, "channelCapacity", "", 10, "channel capacity for concurrency")
	rootCmd.Flags().Int64VarP(&opts.GracePeriodSeconds, "gracePeriodSeconds", "", 30, "grace period to delete target pods")
	rootCmd.Flags().BoolVarP(&opts.TerminateEvicted, "terminateEvicted", "", true, "terminate evicted pods in specified namespaces")
	rootCmd.Flags().Int32VarP(&opts.TerminatingStateMinutes, "terminatingStateMinutes", "", 30, "terminate stucked pods "+
		"in terminating state which are more than that value")
	rootCmd.Flags().BoolVarP(&opts.OneShot, "oneShot", "", true, "specifier to run kube-pod-terminator "+
		"only one time instead of continuously running in the background. should be true if you are using app as CLI.")
	rootCmd.Flags().StringVarP(&opts.BannerFilePath, "bannerFilePath", "", "build/ci/banner.txt",
		"relative path of the banner file")
	rootCmd.Flags().BoolVarP(&opts.VerboseLog, "verbose", "v", false, "verbose output of the logging library (default false)")

	if err := rootCmd.Flags().MarkHidden("bannerFilePath"); err != nil {
		panic("fatal error occured while hiding flag")
	}

	logger = logging.GetLogger()
	logger = logger.With(zap.Bool("inCluster", opts.InCluster))
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "kube-pod-terminator",
	Short:   "Unwanted pod cleaner for Kubernetes(terminating/evicted)",
	Version: ver.GitVersion,
	Long: `On some Kubernetes versions, there is a problem that pods stuck in **Terminating** state on some circumstances. This tool
connects to the **kube-apiserver**, discovers Terminating pods which are in Terminating status and destroys them. This tool can also be
used for Evicted state pods.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(opts.VerboseLog)
		if opts.VerboseLog {
			logging.Atomic.SetLevel(zap.DebugLevel)
		}

		if _, err := os.Stat(opts.BannerFilePath); err == nil {
			bannerBytes, _ := os.ReadFile(opts.BannerFilePath)
			banner.Init(os.Stdout, true, false, strings.NewReader(string(bannerBytes)))
		}

		logger.Info("kube-pod-terminator is started",
			zap.String("appVersion", ver.GitVersion),
			zap.String("goVersion", ver.GoVersion),
			zap.String("goOS", ver.GoOs),
			zap.String("goArch", ver.GoArch),
			zap.String("gitCommit", ver.GitCommit),
			zap.String("buildDate", ver.BuildDate))

		// our application logic starts right here
		kubeConfigPathArr = strings.Split(opts.KubeConfigPaths, ",")
		exitSignal := make(chan os.Signal)
		for _, path := range kubeConfigPathArr {
			go func(p string) {
				logger = logger.With(zap.String("kubeConfigPath", p))
				logger.Info("starting generating clientset for kubeconfig")
				restConfig, err := k8s.GetConfig(p, opts.InCluster)
				if err != nil {
					logger.Fatal("fatal error occurred while getting k8s config", zap.String("error", err.Error()))
				}

				clientSet, err := k8s.GetClientSet(restConfig)
				if err != nil {
					logger.Fatal("fatal error occurred while getting clientset", zap.String("error", err.Error()))
				}

				k8s.Run(opts, clientSet, restConfig.Host)
				if opts.OneShot {
					exitSignal <- syscall.SIGTERM
					return
				}

				ticker := time.NewTicker(time.Duration(opts.TickerIntervalMinutes) * time.Minute)
				for range ticker.C {
					k8s.Run(opts, clientSet, restConfig.Host)
				}
			}(path)
		}

		if opts.OneShot {
			signalCounter := 0
			for range exitSignal {
				signalCounter++
				if signalCounter == len(kubeConfigPathArr) {
					logger.Info("all goroutines sent a SIGTERM, exiting")
					os.Exit(0)
				}
			}
		}

		select {}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
