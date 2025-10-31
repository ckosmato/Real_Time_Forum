 /**
 * UI Manager Module
 * Handles view management, toast notifications, loading states, and utility functions
 */
class UIManager {
    constructor(app) {
        this.app = app;
        this.currentView = 'home';
    }

    /**
     * Initialize UI functionality
     */
    init() {
        this.bindNavigationEvents();
        this.initializeWidget();
    }

    /**
     * Bind navigation-related event listeners
     */
    bindNavigationEvents() {
        // Navigation events
        document.getElementById('home-btn').addEventListener('click', () => this.showView('home'));
        document.getElementById('my-posts-btn').addEventListener('click', () => this.showView('my-posts'));
        document.getElementById('create-post-btn').addEventListener('click', () => this.showView('create-post'));
    }

    /**
     * Initialize online users widget
     */
    initializeWidget() {
        try {
            const widget = document.querySelector('.online-users-widget');
            if (!widget) return;

            const content = widget.querySelector('.widget-content');
            const toggle = widget.querySelector('.widget-toggle');
            const header = widget.querySelector('.widget-header');

            // Initialize closed state
            widget.classList.add('closed');
            if (content) content.classList.add('collapsed');
            if (toggle) toggle.classList.add('collapsed');

            // Bind event listeners
            if (toggle) toggle.addEventListener('click', (e) => this.toggleOnlineUsersWidget(e));
            if (header) header.addEventListener('click', (e) => this.toggleOnlineUsersWidget(e));
        } catch (err) {
            console.error('Error initializing online users widget:', err);
        }
    }

    /**
     * Toggle online users widget
     */
    toggleOnlineUsersWidget(e) {
        if (e && e.stopPropagation) e.stopPropagation();
        
        // Use event delegation to get the widget from the click event's path
        const widget = e.target.closest('.online-users-widget');
        if (!widget) return;
        
        // Find elements once
        const content = widget.querySelector('.widget-content');
        const toggle = widget.querySelector('.widget-toggle');
        const isClosed = widget.classList.contains('closed');

        // Toggle all classes in one batch
        [
            [widget, 'closed'],
            [content, 'collapsed'],
            [toggle, 'collapsed']
        ].forEach(([element, className]) => {
            if (element) element.classList.toggle(className, !isClosed);
        });
    }

    /**
     * Show specific view and hide others
     */
    showView(viewName) {
        // Hide all views
        document.querySelectorAll('.view').forEach(view => {
            view.style.display = 'none';
        });

        // Remove active class from nav buttons
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.classList.remove('active');
        });

        // Show selected view
        document.getElementById(`${viewName}-view`).style.display = 'block';
        
        // Add active class to corresponding nav button
        if (viewName !== 'post') {
            document.getElementById(`${viewName}-btn`).classList.add('active');
        }

        // Hide/show sidebar based on view
        const sidebar = document.querySelector('.sidebar');
        if (viewName === 'post' || viewName === 'create-post' || viewName === 'my-posts') {
            sidebar.style.display = 'none';
        } else {
            sidebar.style.display = 'block';
        }

        this.currentView = viewName;

        // Load data based on view
        if (viewName === 'my-posts') {
            this.app.posts.loadMyPosts();
        }
    }

    /**
     * Get current view
     */
    getCurrentView() {
        return this.currentView;
    }

    /**
     * Show loading spinner
     */
    showLoading() {
        document.getElementById('loading').style.display = 'flex';
    }

    /**
     * Hide loading spinner
     */
    hideLoading() {
        document.getElementById('loading').style.display = 'none';
    }

    /**
     * Show toast notification
     */
    showToast(message, type = 'info', timeout = 4000) {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;

        document.getElementById('toast-container').appendChild(toast);

        setTimeout(() => {
            toast.remove();
        }, timeout);
    }

    /**
     * Render active users in the widget
     */
    renderActiveUsers(users) {
        const container = document.getElementById('active-users-list');
        container.innerHTML = '';
        users.forEach(user => {
            const userDiv = document.createElement('div');
            userDiv.className = 'active-user';
            userDiv.textContent = this.escapeHtml(user.Nickname);
            userDiv.dataset.username = user.Nickname;
            userDiv.addEventListener('click', (e) => {
                e.stopPropagation(); // Prevent widget toggle when clicking user
                this.app.chat.openChatWithUser(user.Nickname);
            });
            container.appendChild(userDiv);
        });
    }

    /**
     * Escape HTML characters to prevent XSS
     */
    escapeHtml(text) {
        // Handle undefined, null, or non-string values
        if (text === undefined || text === null) {
            return '';
        }
        
        // Convert to string if it's not already a string
        text = String(text);
        
        const map = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#039;'
        };
        return text.replace(/[&<>"']/g, m => map[m]);
    }

    /**
     * Format date for display
     */
    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
    }

    /**
     * Clear UI state
     */
    clearState() {
        // Reset current view
        this.currentView = 'home';
        
        // Hide all views
        document.querySelectorAll('.view').forEach(view => {
            view.style.display = 'none';
        });
        
        // Remove active classes
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        
        // Clear active users
        const activeUsersContainer = document.getElementById('active-users-list');
        if (activeUsersContainer) {
            activeUsersContainer.innerHTML = '';
        }
        
        // Hide loading if shown
        this.hideLoading();
    }
}