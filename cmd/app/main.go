package main

import "timeTracker/internal/app"

func main() {
	a := app.NewApp()
	a.Migrate()
	a.ListenAndServe()
}