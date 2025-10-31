 /**
 * Main Forum Application Controller
 * Coordinates between different modules and manages application state
 */
class ForumApp {
    constructor() {
        // Initialize module managers
        this.auth = new AuthManager(this);
        this.ui = new UIManager(this);
        this.posts = new PostsManager(this);
        this.chat = new ChatManager(this);
        
        this.init();
    }

    /**
     * Initialize the application
     */
    init() {
        // Initialize all modules
        this.ui.init();
        this.posts.init();
        this.auth.init();
    }

    /**
     * Load dashboard data and initialize real-time features
     */
    async loadDashboard() {
        this.ui.showLoading();
        
        const sessionId = this.auth.getCookie('session_id');

        try {
            // Show the dashboard UI immediately after successful login
            this.auth.showApp();
            this.ui.showView('home');
            this.loadActiveUsers();

            // Now fetch posts and categories and send session id in headers
            console.log('Fetching dashboard data...');
            const response = await fetch('/dashboard', {
                method: 'GET',
                credentials: 'include', // Changed to 'include' for consistency
                headers: {
                    'X-Session-ID': sessionId
                }
            });
            
            const data = await response.json();
            
            if (response.ok) {
                // Update user information from the server response
                if (data.user) {
                    this.auth.setCurrentUser({
                        nickname: data.user.Nickname,
                    });
                } else {
                    this.clearState();
                    this.auth.showAuth();
                    document.getElementById('login-form').style.display = 'block';
                    document.getElementById('register-form').style.display = 'none'
                }

                this.posts.setCategories(data.categories || []);
                this.posts.setPosts(data.posts || []);
            } else {
                console.error('Failed to load posts:', data.error);
                this.ui.showToast('Failed to load posts. Please refresh the page.', 'error');
                // Still render categories even if the request failed
                this.posts.renderCategories();
            }
        } catch (error) {
            console.error('Dashboard error:', error);
            this.ui.showToast('Failed to load content. Please refresh the page.', 'error');
            // Still render categories even if there's a network error
            this.posts.renderCategories();
        } finally {
            this.ui.hideLoading();
            // Initialize chat after dashboard is loaded
            this.chat.init();
        }
    }

    /**
     * Load active users and display them in the widget
     */
    async loadActiveUsers() {
        this.ui.showLoading();

        try {
            const response = await fetch('/dashboard/active-users', {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'X-Session-ID': this.auth.getCookie('session_id')
                }
            });

            if (response.ok) {
                const data = await response.json();
                this.ui.renderActiveUsers(data.users);
            } else {
                console.error('Failed to load active users');
            }
        } catch (error) {
            console.error('Error loading active users:', error);
        } finally {
            this.ui.hideLoading();
        }
    }

    /**
     * Clear all application state
     */
    clearState() {
        // Clear state in all modules
        this.auth.clearCurrentUser();
        this.ui.clearState();
        this.posts.clearState();
        this.chat.clearState();
    }
}

// Initialize the app
const app = new ForumApp();

// Make app and its methods available globally for onclick handlers
window.app = app;