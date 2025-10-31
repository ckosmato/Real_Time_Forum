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
     * Render active users in the sidebar
     */
    renderActiveUsers(users) {
        const container = document.getElementById('active-users-list');
        const countElement = document.getElementById('online-count');
        
        if (!container) return;

        // Update online count
        if (countElement) {
            countElement.textContent = users.length;
        }

        const isMobile = window.innerWidth <= 768;
        
        // Get current user list to compare for animations
        const currentUsers = Array.from(container.children).map(el => el.dataset.username);
        const newUsers = users.map(user => user.Nickname);
        
        // Find users who joined and left
        const joinedUsers = newUsers.filter(user => !currentUsers.includes(user));
        const leftUsers = currentUsers.filter(user => !newUsers.includes(user));
        
        // Remove leaving users with animation
        leftUsers.forEach(username => {
            const userElement = container.querySelector(`[data-username="${username}"]`);
            if (userElement) {
                userElement.classList.add('leaving');
                setTimeout(() => {
                    if (userElement.parentNode) {
                        userElement.remove();
                    }
                }, 400);
            }
        });
        
        // Clear container for full rebuild (avoiding animation conflicts)
        setTimeout(() => {
            container.innerHTML = '';
            
            users.forEach(user => {
                const userDiv = document.createElement('div');
                userDiv.className = 'active-user';
                userDiv.setAttribute('role', 'button');
                userDiv.setAttribute('tabindex', '0');
                userDiv.setAttribute('aria-label', `Start chat with ${user.Nickname}`);
                
                // Truncate long usernames on mobile
                let displayName = this.escapeHtml(user.Nickname);
                if (isMobile && displayName.length > 12) {
                    displayName = displayName.substring(0, 10) + '...';
                }
                
                userDiv.textContent = displayName;
                userDiv.dataset.username = user.Nickname;
                userDiv.title = this.escapeHtml(user.Nickname);
                
                // Add joining animation for new users
                if (joinedUsers.includes(user.Nickname)) {
                    userDiv.classList.add('joining');
                    // Remove animation class after animation completes
                    setTimeout(() => {
                        userDiv.classList.remove('joining');
                    }, 600);
                }
                
                // Click and keyboard event handlers
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
                
                container.appendChild(userDiv);
            });
            
            // Add scroll indicator on mobile if there are many users
            if (isMobile && users.length > 8) {
                const scrollIndicator = document.createElement('div');
                scrollIndicator.className = 'scroll-indicator';
                scrollIndicator.innerHTML = '<i class="fa-solid fa-chevron-right"></i>';
                container.appendChild(scrollIndicator);
            }
        }, leftUsers.length > 0 ? 450 : 0);
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