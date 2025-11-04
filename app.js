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
            // Load all users immediately to show them even before WebSocket connects
            this.loadAllUsers();
            // Initialize chat after dashboard is loaded - this will get online users via WebSocket
            this.chat.init();
            // Don't load all users immediately - let WebSocket connection establish first
            // All users will be loaded when we receive the initial online users list
        }
    }

    /**
     * Load all users from the server
     */
    async loadAllUsers() {
        console.log('loadAllUsers called');
        const sessionId = this.auth.getCookie('session_id');
        console.log('Session ID:', sessionId);

        try {
            console.log('Fetching all users...');
            const response = await fetch('/dashboard/all-users', {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'X-Session-ID': sessionId
                }
            });
            
            const data = await response.json();
            console.log('All users response:', data);
            
            if (response.ok) {
                console.log('Calling renderAllUsers with:', data.users);
                // Display all users in the UI
                this.ui.renderAllUsers(data.users || []);
            } else {
                console.error('Failed to load all users:', data.error);
            }
        } catch (error) {
            console.error('Error loading all users:', error);
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