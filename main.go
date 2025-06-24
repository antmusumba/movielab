package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// Movie represents a movie or TV show from TMDB
type Movie struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	Rating      float64 `json:"rating"`
	Genre       string  `json:"genre"`
	Type        string  `json:"type"` // movie or tv
}

// WatchlistItem represents an item in the user's watchlist
type WatchlistItem struct {
	ID         int    `json:"id"`
	MovieID    int    `json:"movie_id"`
	Title      string `json:"title"`
	Watched    bool   `json:"watched"`
	AddedAt    string `json:"added_at"`
	PosterPath string `json:"poster_path"`
}

// TMDBResponse represents the structure of a TMDB API response
type TMDBResponse struct {
	Results []struct {
		ID           int     `json:"id"`
		Title        string  `json:"title"`
		Name         string  `json:"name"`
		Overview     string  `json:"overview"`
		PosterPath   string  `json:"poster_path"`
		ReleaseDate  string  `json:"release_date"`
		FirstAirDate string  `json:"first_air_date"`
		VoteAverage  float64 `json:"vote_average"`
		GenreIDs     []int   `json:"genre_ids"`
		MediaType    string  `json:"media_type"`
	} `json:"results"`
	TotalPages int `json:"total_pages"`
}

// OMDBResponse represents the structure of an OMDB API response
type OMDBResponse struct {
	Title          string `json:"Title"`
	Year           string `json:"Year"`
	Plot           string `json:"Plot"`
	IMDBRating     string `json:"imdbRating"`
	RottenTomatoes string `json:"Ratings"`
}

var (
	db         *sql.DB
	tmdbAPIKey string
	omdbAPIKey string
	youtubeAPIKey string
)

// init loads API keys from environment variables
func init() {
	tmdbAPIKey = os.Getenv("TMDB_API_KEY")
	omdbAPIKey = os.Getenv("OMDB_API_KEY")
	youtubeAPIKey = os.Getenv("YOUTUBE_API_KEY")

	if tmdbAPIKey == "" {
		log.Fatal("TMDB_API_KEY environment variable is required")
	}
	if omdbAPIKey == "" {
		log.Fatal("OMDB_API_KEY environment variable is required")
	}
	if youtubeAPIKey == "" {
		log.Fatal("YOUTUBE_API_KEY environment variable is required")
	}
}

// initDB initializes the SQLite database and creates tables if they don't exist
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./movielab.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create tables for watchlist and user preferences
	createTables := `
	CREATE TABLE IF NOT EXISTS watchlist (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		movie_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		watched BOOLEAN DEFAULT FALSE,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		poster_path TEXT,
		media_type TEXT DEFAULT 'movie'
	);
	
	CREATE TABLE IF NOT EXISTS user_preferences (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		preference_key TEXT UNIQUE NOT NULL,
		preference_value TEXT NOT NULL
	);
	`

	_, err = db.Exec(createTables)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()

	// API Routes
	r.HandleFunc("/api/search", searchHandler).Methods("GET")
	r.HandleFunc("/api/trending", trendingHandler).Methods("GET")
	r.HandleFunc("/api/movie/{id}", movieDetailHandler).Methods("GET")
	r.HandleFunc("/api/watchlist", watchlistHandler).Methods("GET", "POST", "DELETE")
	r.HandleFunc("/api/watchlist/{id}", watchlistItemHandler).Methods("PUT", "DELETE")
	r.HandleFunc("/api/recommendations", recommendationsHandler).Methods("GET")
	r.HandleFunc("/api/trailer", youtubeTrailerHandler).Methods("GET")
	r.HandleFunc("/api/trending-trailers", trendingTrailersHandler).Methods("GET")

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/", homeHandler).Methods("GET")

	// CORS middleware for cross-origin requests
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(r)))
}

// homeHandler serves the main HTML page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

// searchHandler handles searching for movies/TV shows via TMDB API
func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Search TMDB API
	url := fmt.Sprintf("https://api.themoviedb.org/3/search/multi?api_key=%s&query=%s&page=%s", tmdbAPIKey, query, page)

	resp, err := http.Get(url)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tmdbResp TMDBResponse
	if err := json.Unmarshal(body, &tmdbResp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Transform TMDB results to Movie structs
	var movies []Movie
	for _, result := range tmdbResp.Results {
		title := result.Title
		if title == "" {
			title = result.Name
		}

		releaseDate := result.ReleaseDate
		if releaseDate == "" {
			releaseDate = result.FirstAirDate
		}

		mediaType := "movie"
		if result.MediaType == "tv" {
			mediaType = "tv"
		}

		movie := Movie{
			ID:          result.ID,
			Title:       title,
			Overview:    result.Overview,
			PosterPath:  result.PosterPath,
			ReleaseDate: releaseDate,
			Rating:      result.VoteAverage,
			Type:        mediaType,
		}
		movies = append(movies, movie)
	}

	// Respond with search results
	response := map[string]interface{}{
		"results":     movies,
		"total_pages": tmdbResp.TotalPages,
		"page":        page,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// trendingHandler returns trending movies or TV shows from TMDB
func trendingHandler(w http.ResponseWriter, r *http.Request) {
	mediaType := r.URL.Query().Get("type")
	if mediaType == "" {
		mediaType = "movie"
	}

	url := fmt.Sprintf("https://api.themoviedb.org/3/trending/%s/week?api_key=%s", mediaType, tmdbAPIKey)

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tmdbResp TMDBResponse
	if err := json.Unmarshal(body, &tmdbResp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var movies []Movie
	for _, result := range tmdbResp.Results {
		title := result.Title
		if title == "" {
			title = result.Name
		}

		releaseDate := result.ReleaseDate
		if releaseDate == "" {
			releaseDate = result.FirstAirDate
		}

		movie := Movie{
			ID:          result.ID,
			Title:       title,
			Overview:    result.Overview,
			PosterPath:  result.PosterPath,
			ReleaseDate: releaseDate,
			Rating:      result.VoteAverage,
			Type:        mediaType,
		}
		movies = append(movies, movie)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

// movieDetailHandler returns detailed info for a specific movie, merging TMDB and OMDB data
func movieDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	movieID := vars["id"]

	// Get TMDB details
	tmdbURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s?api_key=%s&append_to_response=credits", movieID, tmdbAPIKey)

	resp, err := http.Get(tmdbURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var movieDetail map[string]interface{}
	if err := json.Unmarshal(body, &movieDetail); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get OMDB details for additional ratings
	title := movieDetail["title"].(string)
	year := ""
	if releaseDate, ok := movieDetail["release_date"].(string); ok && len(releaseDate) >= 4 {
		year = releaseDate[:4]
	}

	omdbURL := fmt.Sprintf("http://www.omdbapi.com/?t=%s&y=%s&apikey=%s", title, year, omdbAPIKey)

	omdbResp, err := http.Get(omdbURL)
	if err == nil {
		defer omdbResp.Body.Close()
		omdbBody, _ := io.ReadAll(omdbResp.Body)
		var omdbData map[string]interface{}
		if json.Unmarshal(omdbBody, &omdbData) == nil {
			// Merge OMDB data with TMDB data
			movieDetail["omdb"] = omdbData
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movieDetail)
}

// watchlistHandler handles GET, POST, and DELETE for the user's watchlist
func watchlistHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Return all watchlist items
		rows, err := db.Query("SELECT id, movie_id, title, watched, added_at, poster_path FROM watchlist ORDER BY added_at DESC")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var items []WatchlistItem
		for rows.Next() {
			var item WatchlistItem
			err := rows.Scan(&item.ID, &item.MovieID, &item.Title, &item.Watched, &item.AddedAt, &item.PosterPath)
			if err != nil {
				continue
			}
			items = append(items, item)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)

	case "POST":
		// Add a new item to the watchlist
		var item WatchlistItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.Exec("INSERT INTO watchlist (movie_id, title, poster_path) VALUES (?, ?, ?)",
			item.MovieID, item.Title, item.PosterPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		item.ID = int(id)
		item.AddedAt = time.Now().Format("2006-01-02 15:04:05")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)

	case "DELETE":
		// Clear entire watchlist
		_, err := db.Exec("DELETE FROM watchlist")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// watchlistItemHandler handles updating or deleting a specific watchlist item
func watchlistItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	switch r.Method {
	case "PUT":
		// Update watched status of a watchlist item
		var item WatchlistItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE watchlist SET watched = ? WHERE id = ?", item.Watched, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	case "DELETE":
		// Delete a watchlist item by ID
		_, err := db.Exec("DELETE FROM watchlist WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// recommendationsHandler returns movie recommendations based on the user's watchlist
func recommendationsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user's watchlist to generate recommendations
	rows, err := db.Query("SELECT movie_id FROM watchlist LIMIT 5")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movieIDs []string
	for rows.Next() {
		var movieID int
		if err := rows.Scan(&movieID); err != nil {
			continue
		}
		movieIDs = append(movieIDs, strconv.Itoa(movieID))
	}

	if len(movieIDs) == 0 {
		// If no watchlist, return trending movies
		trendingHandler(w, r)
		return
	}

	// Get recommendations based on first movie in watchlist
	movieID := movieIDs[0]
	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s/recommendations?api_key=%s", movieID, tmdbAPIKey)

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tmdbResp TMDBResponse
	if err := json.Unmarshal(body, &tmdbResp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var movies []Movie
	for _, result := range tmdbResp.Results {
		movie := Movie{
			ID:          result.ID,
			Title:       result.Title,
			Overview:    result.Overview,
			PosterPath:  result.PosterPath,
			ReleaseDate: result.ReleaseDate,
			Rating:      result.VoteAverage,
			Type:        "movie",
		}
		movies = append(movies, movie)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

// youtubeTrailerHandler searches YouTube for the official trailer and returns the videoId
func youtubeTrailerHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	year := r.URL.Query().Get("year")
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	query := fmt.Sprintf("%s %s official trailer", title, year)
	apiKey := youtubeAPIKey
	ytURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&type=video&key=%s&maxResults=1", url.QueryEscape(query), apiKey)
	resp, err := http.Get(ytURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	var ytResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ytResp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items, ok := ytResp["items"].([]interface{})
	if !ok || len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	id, ok := items[0].(map[string]interface{})["id"].(map[string]interface{})["videoId"].(string)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"videoId": id})
}

// trendingTrailersHandler returns the top 4 trending movies with their YouTube trailer videoIds
func trendingTrailersHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch trending movies from TMDB
	tmdbURL := fmt.Sprintf("https://api.themoviedb.org/3/trending/movie/week?api_key=%s&page=1", tmdbAPIKey)
	resp, err := http.Get(tmdbURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	var tmdbResp TMDBResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmdbResp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results := []map[string]interface{}{}
	max := 4
	if len(tmdbResp.Results) < max {
		max = len(tmdbResp.Results)
	}
	for i := 0; i < max; i++ {
		movie := tmdbResp.Results[i]
		title := movie.Title
		if title == "" {
			title = movie.Name
		}
		releaseDate := movie.ReleaseDate
		if releaseDate == "" {
			releaseDate = movie.FirstAirDate
		}
		// Search YouTube for trailer
		query := fmt.Sprintf("%s %s official trailer", title, releaseDate[:4])
		ytURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&type=video&key=%s&maxResults=1", url.QueryEscape(query), youtubeAPIKey)
		ytResp, err := http.Get(ytURL)
		if err != nil {
			continue
		}
		var ytData map[string]interface{}
		if err := json.NewDecoder(ytResp.Body).Decode(&ytData); err != nil {
			ytResp.Body.Close()
			continue
		}
		ytResp.Body.Close()
		videoId := ""
		if items, ok := ytData["items"].([]interface{}); ok && len(items) > 0 {
			id, ok := items[0].(map[string]interface{})["id"].(map[string]interface{})["videoId"].(string)
			if ok {
				videoId = id
			}
		}
		results = append(results, map[string]interface{}{
			"title":        title,
			"poster_path":  movie.PosterPath,
			"videoId":      videoId,
			"release_date": releaseDate,
			"overview":     movie.Overview,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
