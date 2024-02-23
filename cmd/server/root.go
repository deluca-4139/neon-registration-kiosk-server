package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var orgId, neonKey, cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&orgId, "orgid", "", "Organization ID; used as the 'username' for Neon requests")
	rootCmd.PersistentFlags().StringVar(&neonKey, "neonkey", "", "API key for making requests to Neon CRM backend")
	viper.BindPFlag("orgid", rootCmd.PersistentFlags().Lookup("orgid"))
	viper.BindPFlag("neonkey", rootCmd.PersistentFlags().Lookup("neonkey"))
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("cobra")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Backend for automated Neon CRM event registration",
	Run: func(cmd *cobra.Command, args []string) {
		orgId = viper.Get("orgId").(string)
		neonKey = viper.Get("neonkey").(string)
		r := chi.NewRouter()

		r.Use(middleware.Logger)
		r.Use(cors.Handler(cors.Options{
			// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
			AllowedOrigins: []string{"https://*", "http://*"},
			// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}))

		// r.Get("/", landingPage)
		// r.Get("/form.js", func(w http.ResponseWriter, r *http.Request) {
		// 	http.ServeFile(w, r, "web/static/form.js")
		// })
		r.Get("/refresh", refreshEvents)
		r.Get("/serverStatus", getServerStatus)
		r.Post("/addEvent", addEvent)
		r.Post("/verify", verifyRegistration)

		http.ListenAndServe(":3000", r)
	},
}
