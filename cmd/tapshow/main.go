package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tapshow/tapshow/internal/config"
	"github.com/tapshow/tapshow/internal/display"
	"github.com/tapshow/tapshow/internal/input"
	"github.com/tapshow/tapshow/internal/privacy"
	"github.com/tapshow/tapshow/internal/processor"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "tapshow",
		Short: "Keystroke visualizer for Wayland",
		Long: `tapshow displays your keystrokes as a minimal overlay window.
Designed for screen recordings, presentations, and live coding.`,
		RunE: run,
	}

	rootCmd.AddCommand(
		configCmd(),
		versionCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	compositor := display.Detect()
	fmt.Printf("Detected compositor: %s\n", compositor)

	backend := display.New()
	fmt.Println("Using GTK window backend")
	showGTKWindowTips(compositor)

	if err := backend.Init(cfg); err != nil {
		return fmt.Errorf("initializing display: %w", err)
	}

	reader := input.NewReader()
	if err := reader.Start(); err != nil {
		return fmt.Errorf("starting input reader: %w", err)
	}
	defer reader.Stop()

	procCfg := processor.Config{
		CombineModifiers: cfg.Behavior.CombineModifiers,
		ShowModifierOnly: cfg.Behavior.ShowModifierOnly,
		ShowHeldKeys:     cfg.Display.ShowHeldKeys,
		HeldKeyTimeout:   cfg.HeldKeyTimeout(),
		HistoryCount:     cfg.Display.HistoryCount,
		ExcludedKeys:     cfg.Behavior.ExcludedKeys,
	}
	proc := processor.New(procCfg)

	go proc.Process(reader.Events())
	defer proc.Stop()

	privacyMonitor := privacy.NewMonitor(cfg.Privacy.PauseOnApps, func(paused bool) {
		backend.SetPaused(paused)
		if paused {
			fmt.Println("Privacy: paused (sensitive app focused)")
		} else {
			fmt.Println("Privacy: resumed")
		}
	})
	privacyMonitor.Start()
	defer privacyMonitor.Stop()

	go func() {
		for event := range proc.Events() {
			backend.Show(event)
			backend.UpdateHistory(proc.History())
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		backend.Stop()
	}()

	fmt.Println("tapshow running. Press Ctrl+C to exit.")
	return backend.Run()
}

func showGTKWindowTips(compositor display.Compositor) {
	switch compositor {
	case display.CompositorKDE:
		fmt.Println("Tip: Right-click the window → More Actions → Keep Above Others")
	case display.CompositorGNOME:
		fmt.Println("Tip: Right-click the window → Always on Top")
	}
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "path",
			Short: "Print the configuration file path",
			RunE: func(cmd *cobra.Command, args []string) error {
				path, err := config.Path()
				if err != nil {
					return err
				}
				fmt.Println(path)
				return nil
			},
		},
		&cobra.Command{
			Use:   "init",
			Short: "Create a default configuration file",
			RunE: func(cmd *cobra.Command, args []string) error {
				path, err := config.Path()
				if err != nil {
					return err
				}

				if _, err := os.Stat(path); err == nil {
					return fmt.Errorf("config already exists at %s", path)
				}

				cfg := config.Default()
				if err := cfg.Save(); err != nil {
					return err
				}

				fmt.Printf("Created config at: %s\n", path)
				return nil
			},
		},
	)

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("tapshow %s\n", version)
		},
	}
}
