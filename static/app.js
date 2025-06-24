// Global variables
let currentPage = 1;
let totalPages = 1;
let searchTimeout;
let currentSection = 'trending';

// DOM elements
const navButtons = document.querySelectorAll('.nav-btn');
const contentSections = document.querySelectorAll('.content-section');
const searchInput = document.getElementById('search-input');
const searchBtn = document.getElementById('search-btn');
const movieModal = document.getElementById('movie-modal');
const modalContent = document.getElementById('movie-detail-content');
const closeModal = document.querySelector('.close');
const loadingOverlay = document.getElementById('loading-overlay');
const clearWatchlistBtn = document.getElementById('clear-watchlist');

// Initialize the app
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
    setupEventListeners();
    loadTrendingContent();
});

function initializeApp() {
    // Load watchlist on startup
    loadWatchlist();
    
    // Load recommendations
    loadRecommendations();
}

function setupEventListeners() {
    // Navigation
    navButtons.forEach(btn => {
        btn.addEventListener('click', () => switchSection(btn.dataset.section));
    });

    // Search 
    searchInput.addEventListener('input', handleSearchInput);
    searchBtn.addEventListener('click', performSearch);
    

    // Modal
    closeModal.addEventListener('click', closeMovieModal);
    window.addEventListener('click', (e) => {
        if (e.target === movieModal) {
            closeMovieModal();
        }
    });

    // Watchlist
    clearWatchlistBtn.addEventListener('click', clearWatchlist);

    // Filter buttons
    document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            loadTrendingContent(btn.dataset.type);
        });
    });
}

// Navigation
function switchSection(sectionName) {
    // Update navigation
    navButtons.forEach(btn => btn.classList.remove('active'));
    document.querySelector(`[data-section="${sectionName}"]`).classList.add('active');

    // Update content sections
    contentSections.forEach(section => section.classList.remove('active'));
    document.getElementById(sectionName).classList.add('active');

    currentSection = sectionName;

    // Load content based on section
    switch(sectionName) {
        case 'trending':
            loadTrendingContent();
            break;
        case 'search':
            // Search section is already loaded
            break;
        case 'watchlist':
            loadWatchlist();
            break;
        case 'recommendations':
            loadRecommendations();
            break;
    }
}

// Search functionality
function handleSearchInput() {
    clearTimeout(searchTimeout);
    const query = searchInput.value.trim();
    
    if (query.length >= 2) {
        searchTimeout = setTimeout(() => {
            performSearch();
        }, 500); // Debounce for 500ms
    } else if (query.length === 0) {
        showSearchPlaceholder();
    }
}

function performSearch() {
    const query = searchInput.value.trim();
    if (query.length < 2) return;

    showLoading('#search-results');
    currentPage = 1;
    
    fetch(`/api/search?q=${encodeURIComponent(query)}&page=${currentPage}`)
        .then(response => response.json())
        .then(data => {
            displaySearchResults(data.results, data.total_pages);
            setupPagination(data.total_pages, query);
        })
        .catch(error => {
            console.error('Search error:', error);
            showError('#search-results', 'Failed to search. Please try again.');
        });
}

function displaySearchResults(movies, totalPages) {
    const container = document.getElementById('search-results');
    
    if (movies.length === 0) {
        container.innerHTML = `
            <div class="search-placeholder">
                <i class="fas fa-search"></i>
                <p>No results found for your search</p>
            </div>
        `;
        return;
    }

    container.innerHTML = movies.map(movie => createMovieCard(movie)).join('');
    setupMovieCardListeners(container);
}

function setupPagination(total, query) {
    const pagination = document.getElementById('search-pagination');
    totalPages = total;
    
    if (total <= 1) {
        pagination.innerHTML = '';
        return;
    }

    let paginationHTML = '';
    
    // Previous button
    paginationHTML += `
        <button ${currentPage === 1 ? 'disabled' : ''} onclick="changePage(${currentPage - 1}, '${query}')">
            <i class="fas fa-chevron-left"></i>
        </button>
    `;

    // Page numbers
    const startPage = Math.max(1, currentPage - 2);
    const endPage = Math.min(total, currentPage + 2);

    for (let i = startPage; i <= endPage; i++) {
        paginationHTML += `
            <button class="${i === currentPage ? 'active' : ''}" onclick="changePage(${i}, '${query}')">
                ${i}
            </button>
        `;
    }

    // Next button
    paginationHTML += `
        <button ${currentPage === total ? 'disabled' : ''} onclick="changePage(${currentPage + 1}, '${query}')">
            <i class="fas fa-chevron-right"></i>
        </button>
    `;

    pagination.innerHTML = paginationHTML;
}

function changePage(page, query) {
    currentPage = page;
    showLoading('#search-results');
    
    fetch(`/api/search?q=${encodeURIComponent(query)}&page=${currentPage}`)
        .then(response => response.json())
        .then(data => {
            displaySearchResults(data.results, data.total_pages);
            setupPagination(data.total_pages, query);
            window.scrollTo({ top: 0, behavior: 'smooth' });
        })
        .catch(error => {
            console.error('Page change error:', error);
            showError('#search-results', 'Failed to load page. Please try again.');
        });
}

function showSearchPlaceholder() {
    document.getElementById('search-results').innerHTML = `
        <div class="search-placeholder">
            <i class="fas fa-search"></i>
            <p>Search for your favorite movies and TV shows</p>
        </div>
    `;
    document.getElementById('search-pagination').innerHTML = '';
}

// Trending content
function loadTrendingContent(type = 'movie') {
    showLoading('#trending-grid');
    
    fetch(`/api/trending?type=${type}`)
        .then(response => response.json())
        .then(movies => {
            displayMovies('#trending-grid', movies);
        })
        .catch(error => {
            console.error('Trending error:', error);
            showError('#trending-grid', 'Failed to load trending content.');
        });
}

// Watchlist functionality
function loadWatchlist() {
    showLoading('#watchlist-grid');
    
    fetch('/api/watchlist')
        .then(response => response.json())
        .then(items => {
            displayWatchlist(items);
        })
        .catch(error => {
            console.error('Watchlist error:', error);
            showError('#watchlist-grid', 'Failed to load watchlist.');
        });
}

function displayWatchlist(items) {
    const container = document.getElementById('watchlist-grid');
    
    if (items.length === 0) {
        container.innerHTML = `
            <div class="empty-watchlist">
                <i class="fas fa-bookmark"></i>
                <p>Your watchlist is empty</p>
                <p>Add movies and TV shows to keep track of what you want to watch</p>
            </div>
        `;
        return;
    }

    container.innerHTML = items.map(item => createWatchlistCard(item)).join('');
    setupWatchlistCardListeners(container);
}

function createWatchlistCard(item) {
    const posterUrl = item.poster_path 
        ? `https://image.tmdb.org/t/p/w500${item.poster_path}`
        : '/static/placeholder-poster.jpg';
    
    return `
        <div class="movie-card ${item.watched ? 'watched' : ''}" data-id="${item.id}">
            <img src="${posterUrl}" alt="${item.title}" class="movie-poster" onerror="this.src='/static/placeholder-poster.jpg'">
            <div class="movie-info">
                <h3 class="movie-title">${item.title}</h3>
                <div class="movie-meta">
                    <span>Added: ${formatDate(item.added_at)}</span>
                </div>
                <div class="movie-actions">
                    <button class="action-btn ${item.watched ? 'btn-secondary' : 'btn-primary'}" onclick="toggleWatched(${item.id}, ${!item.watched})">
                        <i class="fas fa-${item.watched ? 'eye-slash' : 'eye'}"></i>
                        ${item.watched ? 'Mark Unwatched' : 'Mark Watched'}
                    </button>
                    <button class="action-btn btn-danger" onclick="removeFromWatchlist(${item.id})">
                        <i class="fas fa-trash"></i>
                        Remove
                    </button>
                </div>
            </div>
        </div>
    `;
}

function addToWatchlist(movie) {
    const watchlistItem = {
        movie_id: movie.id,
        title: movie.title,
        poster_path: movie.poster_path
    };

    fetch('/api/watchlist', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(watchlistItem)
    })
    .then(response => response.json())
    .then(item => {
        showNotification('Added to watchlist!', 'success');
        if (currentSection === 'watchlist') {
            loadWatchlist();
        }
    })
    .catch(error => {
        console.error('Add to watchlist error:', error);
        showNotification('Failed to add to watchlist.', 'error');
    });
}

function removeFromWatchlist(id) {
    fetch(`/api/watchlist/${id}`, {
        method: 'DELETE'
    })
    .then(() => {
        showNotification('Removed from watchlist!', 'success');
        loadWatchlist();
    })
    .catch(error => {
        console.error('Remove from watchlist error:', error);
        showNotification('Failed to remove from watchlist.', 'error');
    });
}

function toggleWatched(id, watched) {
    fetch(`/api/watchlist/${id}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ watched })
    })
    .then(() => {
        showNotification(watched ? 'Marked as watched!' : 'Marked as unwatched!', 'success');
        loadWatchlist();
    })
    .catch(error => {
        console.error('Toggle watched error:', error);
        showNotification('Failed to update status.', 'error');
    });
}

function clearWatchlist() {
    if (!confirm('Are you sure you want to clear your entire watchlist?')) {
        return;
    }

    fetch('/api/watchlist', {
        method: 'DELETE'
    })
    .then(() => {
        showNotification('Watchlist cleared!', 'success');
        loadWatchlist();
    })
    .catch(error => {
        console.error('Clear watchlist error:', error);
        showNotification('Failed to clear watchlist.', 'error');
    });
}

// Recommendations
function loadRecommendations() {
    showLoading('#recommendations-grid');
    
    fetch('/api/recommendations')
        .then(response => response.json())
        .then(movies => {
            displayMovies('#recommendations-grid', movies);
        })
        .catch(error => {
            console.error('Recommendations error:', error);
            showError('#recommendations-grid', 'Failed to load recommendations.');
        });
}

// Movie display functions
function displayMovies(containerId, movies) {
    const container = document.querySelector(containerId);
    
    if (movies.length === 0) {
        container.innerHTML = '<div class="loading">No content available.</div>';
        return;
    }

    container.innerHTML = movies.map(movie => createMovieCard(movie)).join('');
    setupMovieCardListeners(container);
}

function createMovieCard(movie) {
    const posterUrl = movie.poster_path 
        ? `https://image.tmdb.org/t/p/w500${movie.poster_path}`
        : '/static/placeholder-poster.jpg';
    
    return `
        <div class="movie-card" data-id="${movie.id}" data-type="${movie.type}">
            <img src="${posterUrl}" alt="${movie.title}" class="movie-poster" onerror="this.src='/static/placeholder-poster.jpg'">
            <div class="movie-info">
                <h3 class="movie-title">${movie.title}</h3>
                <div class="movie-meta">
                    <span>${movie.release_date ? formatDate(movie.release_date) : 'Unknown'}</span>
                    <div class="movie-rating">
                        <i class="fas fa-star"></i>
                        <span>${movie.rating ? movie.rating.toFixed(1) : 'N/A'}</span>
                    </div>
                </div>
                <p class="movie-overview">${truncateText(movie.overview, 100)}</p>
                <div class="movie-actions">
                    <button class="action-btn btn-primary" onclick="showMovieDetails(${movie.id}, '${movie.type}')">
                        <i class="fas fa-info-circle"></i>
                        Details
                    </button>
                    <button class="action-btn btn-secondary" onclick="addToWatchlist(${JSON.stringify(movie).replace(/"/g, '&quot;')})">
                        <i class="fas fa-bookmark"></i>
                        Add to List
                    </button>
                </div>
            </div>
        </div>
    `;
}

function setupMovieCardListeners(container) {
    container.querySelectorAll('.movie-card').forEach(card => {
        card.addEventListener('click', (e) => {
            if (!e.target.closest('.action-btn')) {
                const id = card.dataset.id;
                const type = card.dataset.type;
                showMovieDetails(id, type);
            }
        });
    });
}

function setupWatchlistCardListeners(container) {
    // Event listeners are handled by onclick attributes in the card HTML
}

// Movie details modal
function showMovieDetails(id, type) {
    showLoadingOverlay();
    movieModal.style.display = 'block';
    
    fetch(`/api/movie/${id}`)
        .then(response => response.json())
        .then(movie => {
            displayMovieDetails(movie);
            hideLoadingOverlay();
        })
        .catch(error => {
            console.error('Movie details error:', error);
            modalContent.innerHTML = '<div class="loading">Failed to load movie details.</div>';
            hideLoadingOverlay();
        });
}

function displayMovieDetails(movie) {
    const posterUrl = movie.poster_path 
        ? `https://image.tmdb.org/t/p/original${movie.poster_path}`
        : '/static/placeholder-poster.jpg';
    
    const genres = movie.genres ? movie.genres.map(g => g.name).join(', ') : '';
    const cast = movie.credits?.cast ? movie.credits.cast.slice(0, 5).map(c => c.name).join(', ') : '';
    
    let ratingsHTML = '';
    if (movie.omdb && movie.omdb.Ratings) {
        ratingsHTML = `
            <div class="ratings-section">
                <h3>Ratings</h3>
                <div class="ratings-grid">
                    ${movie.omdb.Ratings.map(rating => `
                        <div class="rating-item">
                            <div class="rating-source">${rating.Source}</div>
                            <div class="rating-value">${rating.Value}</div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    modalContent.innerHTML = `
        <div class="movie-detail">
            <img src="${posterUrl}" alt="${movie.title}" class="movie-detail-poster" onerror="this.src='/static/placeholder-poster.jpg'">
            <div class="movie-detail-info">
                <h1>${movie.title}</h1>
                <div class="movie-detail-meta">
                    <div class="meta-item">
                        <i class="fas fa-calendar"></i>
                        <span>${movie.release_date || 'Unknown'}</span>
                    </div>
                    <div class="meta-item">
                        <i class="fas fa-clock"></i>
                        <span>${movie.runtime ? `${movie.runtime} min` : 'Unknown'}</span>
                    </div>
                    <div class="meta-item">
                        <i class="fas fa-star"></i>
                        <span>${movie.vote_average ? movie.vote_average.toFixed(1) : 'N/A'}</span>
                    </div>
                </div>
                ${genres ? `<div class="meta-item"><i class="fas fa-tags"></i><span>${genres}</span></div>` : ''}
                <div class="movie-detail-overview">
                    <h3>Overview</h3>
                    <p>${movie.overview || 'No overview available.'}</p>
                </div>
                ${cast ? `
                    <div class="movie-detail-cast">
                        <h3>Cast</h3>
                        <p>${cast}</p>
                    </div>
                ` : ''}
                ${ratingsHTML}
                <div class="movie-actions">
                    <button class="action-btn btn-primary" onclick="addToWatchlist(${JSON.stringify({
                        id: movie.id,
                        title: movie.title,
                        poster_path: movie.poster_path
                    }).replace(/"/g, '&quot;')})">
                        <i class="fas fa-bookmark"></i>
                        Add to Watchlist
                    </button>
                </div>
            </div>
        </div>
    `;
}

function closeMovieModal() {
    movieModal.style.display = 'none';
}

// Utility functions
function showLoading(containerId) {
    document.querySelector(containerId).innerHTML = '<div class="loading">Loading...</div>';
}

function showError(containerId, message) {
    document.querySelector(containerId).innerHTML = `
        <div class="search-placeholder">
            <i class="fas fa-exclamation-triangle"></i>
            <p>${message}</p>
        </div>
    `;
}

function showLoadingOverlay() {
    loadingOverlay.style.display = 'flex';
}

function hideLoadingOverlay() {
    loadingOverlay.style.display = 'none';
}

function showNotification(message, type = 'info') {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <i class="fas fa-${type === 'success' ? 'check-circle' : type === 'error' ? 'exclamation-circle' : 'info-circle'}"></i>
        <span>${message}</span>
    `;
    
    // Add styles
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: ${type === 'success' ? '#27ae60' : type === 'error' ? '#e74c3c' : '#3498db'};
        color: white;
        padding: 1rem 1.5rem;
        border-radius: 8px;
        box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
        z-index: 3000;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        animation: slideIn 0.3s ease;
    `;
    
    document.body.appendChild(notification);
    
    // Remove after 3 seconds
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }, 3000);
}

function truncateText(text, maxLength) {
    if (!text) return '';
    return text.length > maxLength ? text.substring(0, maxLength) + '...' : text;
}

function formatDate(dateString) {
    if (!dateString) return 'Unknown';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'short', 
        day: 'numeric' 
    });
}

// Add CSS animations for notifications
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
    
    @keyframes slideOut {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100%); opacity: 0; }
    }
`;
document.head.appendChild(style); 