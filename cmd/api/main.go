// @title          Agentic Caddie API
// @version        1.0
// @description    Golf caddie backend — players, rounds, shots and course data.
// @host           localhost:3000
// @BasePath       /
package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sbedford/agentic-caddie/internal/config"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/handlers"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/sbedford/agentic-caddie/cmd/api/docs"
	_ "modernc.org/sqlite"
)

func main() {
	cfg := config.Load()
	database, err := sql.Open("sqlite", cfg.DSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	queries := db.New(database)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Mount("/players", playerRoutes(queries))
	r.Mount("/courses", courseRoutes(queries))
	r.Mount("/tees", teeRoutes(queries))
	r.Mount("/course-holes", courseHoleRoutes(queries))
	r.Mount("/tee-holes", teeHoleRoutes(queries))
	r.Mount("/pois", poiRoutes(queries))
	r.Mount("/clubs", clubRoutes(queries))
	r.Mount("/rounds", roundRoutes(queries))
	r.Mount("/holes", holeRoutes(queries))
	r.Mount("/shots", shotRoutes(queries, database))
	r.Mount("/commentary", commentaryRoutes(queries))
	r.Mount("/vocabulary", vocabularyRoutes(queries))

	log.Println("Server listening on :3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func playerRoutes(queries *db.Queries) chi.Router {
	h := handlers.PlayersHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/", h.ListPlayers)
	r.Post("/", h.CreatePlayer)
	r.Get("/name/{name}", h.GetPlayerByName)
	r.Get("/{id}", h.GetPlayer)
	r.Patch("/{id}/handicap", h.UpdateHandicap)
	r.Delete("/{id}", h.DeletePlayer)
	return r
}

func courseRoutes(queries *db.Queries) chi.Router {
	h := handlers.CoursesHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/", h.ListCourses)
	r.Post("/", h.CreateCourse)
	r.Get("/name/{name}", h.GetCourseByName)
	r.Get("/{id}", h.GetCourseByID)
	r.Put("/{id}", h.UpdateCourse)
	r.Delete("/{id}", h.DeleteCourse)
	return r
}

func teeRoutes(queries *db.Queries) chi.Router {
	h := handlers.TeesHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/", h.ListTees)
	r.Post("/", h.CreateTee)
	r.Get("/course/{courseId}", h.GetTeesByCourse)
	r.Get("/course/{courseId}/{name}", h.GetTeeByCourseAndName)
	r.Get("/{id}", h.GetTeeByID)
	r.Put("/{id}", h.UpdateTee)
	r.Delete("/{id}", h.DeleteTee)
	return r
}

func courseHoleRoutes(queries *db.Queries) chi.Router {
	h := handlers.CourseHolesHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/course/{courseId}", h.ListCourseHoles)
	r.Get("/course/{courseId}/number/{holeNumber}", h.GetCourseHoleByCourseAndNumber)
	r.Post("/", h.CreateCourseHole)
	r.Get("/{id}", h.GetCourseHoleByID)
	r.Patch("/{id}/coordinates", h.UpdateCourseHoleCoordinates)
	r.Delete("/{id}", h.DeleteCourseHole)
	return r
}

func teeHoleRoutes(queries *db.Queries) chi.Router {
	h := handlers.TeeHolesHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/tee/{teeId}", h.ListTeeHoles)
	r.Get("/hole/{courseHoleId}/tee/{teeId}", h.GetTeeHoleByHoleAndTee)
	r.Post("/", h.CreateTeeHole)
	r.Get("/{id}", h.GetTeeHoleByID)
	r.Put("/{id}", h.UpdateTeeHole)
	r.Delete("/{id}", h.DeleteTeeHole)
	return r
}

func poiRoutes(queries *db.Queries) chi.Router {
	h := handlers.POIsHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/hole/{courseHoleId}", h.ListPOIsByHole)
	r.Get("/hole/{courseHoleId}/tee/{tee}", h.ListPOIsByHoleAndTee)
	r.Delete("/hole/{courseHoleId}", h.DeletePOIsByHole)
	r.Post("/", h.CreatePOI)
	r.Get("/{id}", h.GetPOIByID)
	r.Put("/{id}", h.UpdatePOI)
	r.Delete("/{id}", h.DeletePOI)
	return r
}

func clubRoutes(queries *db.Queries) chi.Router {
	h := handlers.ClubsHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/player/{playerId}", h.ListClubsByPlayer)
	r.Get("/player/{playerId}/active", h.ListActiveClubsByPlayer)
	r.Get("/player/{playerId}/name/{clubName}", h.GetClubByPlayerAndName)
	r.Get("/player/{playerId}/name/{clubName}/date/{date}", h.GetClubByPlayerNameAndDate)
	r.Patch("/player/{playerId}/name/{clubName}/retire", h.RetireClub)
	r.Post("/", h.CreateClub)
	r.Get("/{id}", h.GetClubByID)
	r.Patch("/{id}/distances", h.UpdateClubDistances)
	r.Delete("/{id}", h.DeleteClub)
	return r
}

func roundRoutes(queries *db.Queries) chi.Router {
	h := handlers.RoundsHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/", h.ListRounds)
	r.Get("/player/{playerId}", h.ListRoundsByPlayer)
	r.Get("/player/{playerId}/course/{courseId}", h.ListRoundsByPlayerAndCourse)
	r.Get("/player/{playerId}/date/{date}", h.GetRoundByPlayerAndDate)
	r.Post("/", h.CreateRound)
	r.Get("/{id}", h.GetRoundByID)
	r.Patch("/{id}/totals", h.UpdateRoundTotals)
	r.Delete("/{id}", h.DeleteRound)
	return r
}

func holeRoutes(queries *db.Queries) chi.Router {
	h := handlers.HolesHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/round/{roundId}", h.ListHolesByRound)
	r.Get("/round/{roundId}/number/{holeNumber}", h.GetHoleByRoundAndNumber)
	r.Post("/", h.CreateHole)
	r.Get("/{id}", h.GetHoleByID)
	r.Put("/{id}", h.UpdateHole)
	r.Delete("/{id}", h.DeleteHole)
	return r
}

func shotRoutes(queries *db.Queries, database *sql.DB) chi.Router {
	h := handlers.ShotsHandler{Queries: queries, DB: database}
	r := chi.NewRouter()
	r.Get("/hole/{holeId}", h.ListShotsByHole)
	r.Get("/hole/{holeId}/type/{shotType}", h.ListShotsByHoleAndType)
	r.Get("/hole/{holeId}/number/{shotNumber}", h.GetShotByHoleAndNumber)
	r.Post("/hole/{holeId}/reorder", h.ReorderShots)
	r.Delete("/hole/{holeId}", h.DeleteShotsByHole)
	r.Post("/", h.CreateShot)
	r.Get("/{id}", h.GetShotByID)
	r.Put("/{id}", h.UpdateShot)
	r.Delete("/{id}", h.DeleteShot)
	return r
}

func vocabularyRoutes(queries *db.Queries) chi.Router {
	h := handlers.VocabularyHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/", h.GetAllVocabulary)
	r.Post("/", h.CreateVocabularyEntry)
	r.Get("/{domain}", h.GetVocabularyByDomain)
	r.Put("/{domain}/{value}", h.UpdateVocabularyEntry)
	r.Delete("/{domain}/{value}", h.DeleteVocabularyEntry)
	return r
}

func commentaryRoutes(queries *db.Queries) chi.Router {
	h := handlers.CommentaryHandler{Queries: queries}
	r := chi.NewRouter()
	r.Get("/scope/{scope}/{scopeId}", h.ListCommentaryByScope)
	r.Get("/scope/{scope}/{scopeId}/latest", h.GetLatestCommentaryByScope)
	r.Delete("/scope/{scope}/{scopeId}", h.DeleteCommentaryByScope)
	r.Post("/", h.CreateCommentary)
	r.Get("/{id}", h.GetCommentaryByID)
	r.Delete("/{id}", h.DeleteCommentary)
	return r
}
