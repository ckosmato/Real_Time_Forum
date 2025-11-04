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
        this.handleResponsiveLayout();
        this.bindResizeEvents();
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
     * Handle responsive layout adjustments
     */
    handleResponsiveLayout() {
        const updateLayout = () => {
            const screenSize = this.getScreenSize();
            const isMobile = screenSize === 'mobile' || screenSize === 'mobile-small';
            const isTablet = screenSize === 'tablet';
            const isTouch = this.isTouchDevice();
            const onlineUsersSidebar = document.querySelector('.online-users-sidebar');
            
            if (onlineUsersSidebar) {
                // Remove all responsive classes first
                onlineUsersSidebar.classList.remove('mobile-layout', 'tablet-layout', 'floating');
                
                // Add appropriate classes
                onlineUsersSidebar.classList.toggle('mobile-layout', isMobile);
                onlineUsersSidebar.classList.toggle('tablet-layout', isTablet);
                
                // Adjust online users display based on screen size and orientation
                if (isMobile && window.innerHeight < window.innerWidth) {
                    // Landscape mobile - make it floating
                    onlineUsersSidebar.classList.add('floating');
                    onlineUsersSidebar.style.position = 'fixed';
                } else if (isMobile) {
                    // Portrait mobile - keep it in flow
                    onlineUsersSidebar.style.position = 'static';
                }
                
                // Add touch-specific styling
                if (isTouch) {
                    onlineUsersSidebar.classList.add('touch-device');
                }
            }
            
            // Update document body with screen size class for global styling
            document.body.className = document.body.className.replace(/screen-\w+/g, '');
            document.body.classList.add(`screen-${screenSize}`);
        };
        
        updateLayout();
    }

    /**
     * Bind window resize events for responsive behavior
     */
    bindResizeEvents() {
        let resizeTimeout;
        window.addEventListener('resize', () => {
            clearTimeout(resizeTimeout);
            resizeTimeout = setTimeout(() => {
                this.handleResponsiveLayout();
            }, 250);
        });
        
        // Handle orientation changes on mobile devices
        window.addEventListener('orientationchange', () => {
            setTimeout(() => {
                this.handleResponsiveLayout();
            }, 500); // Give time for orientation change to complete
        });
    }

    /**
     * Render all users with their online status in the sidebar
     */
    renderAllUsers(users) {
        const container = document.getElementById('active-users-list');
        const countElement = document.getElementById('online-count');
        
        if (!container) return;

        // Get online users from the websocket hub
        const onlineUsers = this.app.chat.getOnlineUsers ? this.app.chat.getOnlineUsers() : [];
        
        // Update online count
        if (countElement) {
            countElement.textContent = onlineUsers.length;
        }

        const isMobile = window.innerWidth <= 768;
        
        // Clear container
        container.innerHTML = '';
        
        // Sort users: online users first, then offline users, both alphabetically
        const sortedUsers = [...users].sort((a, b) => {
            const aIsOnline = onlineUsers.includes(a.Nickname);
            const bIsOnline = onlineUsers.includes(b.Nickname);
            
            // First sort by online status (online first)
            if (aIsOnline && !bIsOnline) return -1;
            if (!aIsOnline && bIsOnline) return 1;
            
            // Then sort alphabetically by nickname
            return a.Nickname.localeCompare(b.Nickname);
        });
        
        sortedUsers.forEach(user => {
            const isOnline = onlineUsers.includes(user.Nickname);
            
            const userDiv = document.createElement('div');
            userDiv.className = `all-user ${isOnline ? 'online' : 'offline'}`;
            userDiv.setAttribute('role', 'button');
            userDiv.setAttribute('tabindex', '0');
            userDiv.setAttribute('aria-label', `${isOnline ? 'Online' : 'Offline'} - ${user.Nickname}`);
            
            // Create user content with online indicator
            const userContent = document.createElement('div');
            userContent.className = 'user-content';
            
            // Online status dot
            const statusDot = document.createElement('div');
            statusDot.className = `status-dot ${isOnline ? 'online' : 'offline'}`;
            statusDot.setAttribute('aria-hidden', 'true');
            
            // Username
            const username = document.createElement('span');
            username.className = 'username';
            
            // Truncate long usernames on mobile
            let displayName = this.escapeHtml(user.Nickname);
            if (isMobile && displayName.length > 10) {
                displayName = displayName.substring(0, 8) + '...';
            }
            
            username.textContent = displayName;
            
            // Assemble the user element
            userContent.appendChild(statusDot);
            userContent.appendChild(username);
            userDiv.appendChild(userContent);
            
            userDiv.dataset.username = user.Nickname;
            userDiv.title = `${this.escapeHtml(user.Nickname)} (${isOnline ? 'Online' : 'Offline'})`;
            
            // Click and keyboard event handlers - only for online users
            if (isOnline) {
                const handleUserInteraction = (e) => {
                    e.stopPropagation();
                    this.app.chat.openChatWithUser(user.Nickname);
                };
                
                userDiv.addEventListener('click', handleUserInteraction);
                userDiv.addEventListener('keydown', (e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        handleUserInteraction(e);
                    }
                });
                
                // Add hover effect for clickable users
                userDiv.style.cursor = 'pointer';
            } else {
                // Disable interaction for offline users
                userDiv.style.cursor = 'default';
                userDiv.style.opacity = '0.6';
            }
            
            container.appendChild(userDiv);
        });
        
        // Add scroll indicator on mobile if there are many users
        if (isMobile && users.length > 8) {
            const scrollIndicator = document.createElement('div');
            scrollIndicator.className = 'scroll-indicator';
            scrollIndicator.innerHTML = '<i class="fa-solid fa-chevron-right"></i>';
            container.appendChild(scrollIndicator);
        }
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

        // Hide/show sidebars based on view
        const leftSidebar = document.querySelector('.sidebar');
        const rightSidebar = document.querySelector('.online-users-sidebar');
        
        if (viewName === 'post' || viewName === 'create-post' || viewName === 'my-posts') {
            leftSidebar.style.display = 'none';
        } else {
            leftSidebar.style.display = 'block';
        }
        
        // Always show the online users sidebar
        if (rightSidebar) {
            rightSidebar.style.display = 'block';
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
     * Check if device supports touch
     */
    isTouchDevice() {
        return 'ontouchstart' in window || navigator.maxTouchPoints > 0;
    }

    /**
     * Get current screen size category
     */
    getScreenSize() {
        const width = window.innerWidth;
        if (width <= 480) return 'mobile-small';
        if (width <= 768) return 'mobile';
        if (width <= 992) return 'tablet';
        if (width <= 1200) return 'desktop-small';
        return 'desktop';
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
        
        // Clear online count
        const countElement = document.getElementById('online-count');
        if (countElement) {
            countElement.textContent = '0';
        }
        
        // Hide loading if shown
        this.hideLoading();
    }
}