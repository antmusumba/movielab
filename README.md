# MovieLab - Movie & TV Show Discovery Web App

A comprehensive entertainment discovery platform built with Go, SQLite, JavaScript, HTML, and CSS. Users can search for movies and TV shows, view detailed information, manage personal watchlists, and discover trending content.

## 🎬 Features

### Core Functionality
- **Search**: Real-time search for movies and TV shows with debounced input
- **Trending Content**: View popular movies and TV shows from TMDB
- **Watchlist Management**: Add/remove titles and mark as watched
- **Recommendations**: Personalized recommendations based on watchlist preferences
- **Detailed Views**: Comprehensive movie/show information with ratings from multiple sources
- **Trailer Integration**: Watch official trailers via YouTube

### Technical Features
- **Responsive Design**: Mobile and desktop optimized
- **API Integration**: TMDB, OMDB, and YouTube APIs for comprehensive data
- **Pagination**: Efficient browsing of search and trending results
- **Error Handling**: Graceful handling of API errors and loading states
- **Caching**: Local storage for user preferences
- **Rate Limiting**: Proper API rate limit handling

## 🛠️ Technology Stack

- **Backend**: Go with Gorilla Mux router
- **Database**: SQLite for local data storage
- **Frontend**: Vanilla JavaScript, HTML5, CSS3
- **APIs**: TMDB (The Movie Database), OMDB (Open Movie Database), YouTube Data API
- **Styling**: Modern CSS with Flexbox/Grid and responsive design

## 📋 Prerequisites

- Go 1.21 or higher
- TMDB API key
- OMDB API key
- YouTube Data API key

## 🚀 Setup Instructions

### 1. Clone the Repository
```bash
git clone <repository-url>
cd movielab
```

### 2. Get API Keys

#### TMDB API Key
1. Visit [The Movie Database](https://www.themoviedb.org/)
2. Create an account and go to Settings > API
3. Request an API key for developer use
4. Copy your API key

#### OMDB API Key
1. Visit [OMDB API](http://www.omdbapi.com/)
2. Request a free API key
3. Copy your API key

#### YouTube Data API Key
1. Visit [Google Cloud Console](https://console.developers.google.com/)
2. Create a project and enable the **YouTube Data API v3**
3. Create an API key
4. Copy your API key

### 3. Set Environment Variables
```bash
export TMDB_API_KEY="your_tmdb_api_key_here"
export OMDB_API_KEY="your_omdb_api_key_here"
export YOUTUBE_API_KEY="your_youtube_api_key_here"
```

### 4. Install Dependencies
```bash
go mod tidy
```

### 5. Run the Application
```bash
go run main.go
```

The application will be available at `http://localhost:8080`

## 📁 Project Structure

```
movielab/
├── main.go              # Go server with API endpoints
├── go.mod               # Go module dependencies
├── static/              # Frontend assets
│   ├── index.html       # Main HTML file
│   ├── styles.css       # CSS styling
│   └── app.js          # JavaScript functionality
├── movielab.db         # SQLite database (created automatically)
└── README.md           # This file
```

## 🔌 API Endpoints

### Search
- `GET /api/search?q={query}&page={page}` - Search movies and TV shows

### Trending
- `GET /api/trending?type={movie|tv}&page={page}` - Get trending content (paginated)

### Movie Details
- `GET /api/movie/{id}` - Get detailed movie information

### Watchlist
- `GET /api/watchlist` - Get user's watchlist
- `POST /api/watchlist` - Add item to watchlist
- `DELETE /api/watchlist` - Clear entire watchlist
- `PUT /api/watchlist/{id}` - Update watchlist item
- `DELETE /api/watchlist/{id}` - Remove item from watchlist

### Recommendations
- `GET /api/recommendations` - Get personalized recommendations

### Trailer
- `GET /api/trailer?title={title}&year={year}` - Get YouTube videoId for the official trailer

## 🎨 Features in Detail

### Search Functionality
- Real-time search with 500ms debouncing
- Pagination support for large result sets
- Genre and year filtering options
- Search across movies, TV shows, and actors

### Watchlist Management
- Add movies/shows to personal watchlist
- Mark items as watched/unwatched
- Remove items from watchlist
- Clear entire watchlist
- Persistent storage in SQLite database

### Movie Details
- Comprehensive movie information
- Cast and crew details
- Multiple rating sources (IMDB, Rotten Tomatoes, TMDB)
- High-quality poster images
- Release date and runtime information
- **Official trailer embedded from YouTube**

### Trending Content
- Weekly trending movies and TV shows
- Filter by content type (movie/TV)
- Real-time data from TMDB
- **Pagination (4 per page)**

### Recommendations
- Personalized recommendations based on watchlist
- Fallback to trending content if no watchlist
- Smart algorithm using TMDB recommendations API

## 🎯 Technical Implementation

### Backend (Go)
- RESTful API design with Gorilla Mux
- SQLite database for local storage
- Proper error handling and HTTP status codes
- CORS support for cross-origin requests
- Environment variable configuration
- **YouTube Data API integration for trailers**

### Frontend (JavaScript)
- Vanilla JavaScript with modern ES6+ features
- Modular code organization
- Event-driven architecture
- Responsive UI with loading states
- Local storage for user preferences
- **Trailer embedding in movie detail modal**

### Database Schema
```sql
-- Watchlist table
CREATE TABLE watchlist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    movie_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    watched BOOLEAN DEFAULT FALSE,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    poster_path TEXT,
    media_type TEXT DEFAULT 'movie'
);

-- User preferences table
CREATE TABLE user_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    preference_key TEXT UNIQUE NOT NULL,
    preference_value TEXT NOT NULL
);
```

## 🔧 Configuration

### Environment Variables
- `TMDB_API_KEY`: Your TMDB API key (required)
- `OMDB_API_KEY`: Your OMDB API key (required)
- `YOUTUBE_API_KEY`: Your YouTube Data API key (required)

### Database
The SQLite database (`movielab.db`) is created automatically on first run.

## 🚀 Deployment

### Local Development
```bash
go run main.go
```

### Production Build
```bash
go build -o movielab main.go
./movielab
```

### Docker (Optional)
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o movielab main.go
EXPOSE 8080
CMD ["./movielab"]
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📝 License

This project is open source and available under the [MIT License](LICENSE).

## 🙏 Acknowledgments

- [The Movie Database (TMDB)](https://www.themoviedb.org/) for movie and TV show data
- [OMDB API](http://www.omdbapi.com/) for additional ratings and plot information
- [YouTube Data API](https://developers.google.com/youtube/v3) for trailers
- [Font Awesome](https://fontawesome.com/) for icons
- [Google Fonts](https://fonts.google.com/) for typography

## 🐛 Troubleshooting

### Common Issues

1. **API Key Errors**: Ensure all API keys are set correctly
2. **Database Errors**: Check file permissions for SQLite database creation
3. **Port Conflicts**: Change the port in `main.go` if 8080 is already in use
4. **CORS Issues**: The server includes CORS headers, but check browser console for errors

### Debug Mode
Add debug logging by setting the log level in the Go application.

## 📊 Performance

- API responses are optimized for speed
- Images are served from TMDB CDN
- Database queries are indexed for efficiency
- Frontend includes debouncing for search input
- Loading states provide smooth user experience

---

**MovieLab** - Discover amazing movies and TV shows with style! 🎬✨ 