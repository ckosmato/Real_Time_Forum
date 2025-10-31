 /**
 * Authentication Module
 * Handles user login, registration, logout, and session management
 */
class AuthManager {
    constructor(app) {
        this.app = app;
        this.currentUser = null;
    }

    /**
     * Initialize authentication event listeners
     */
    init() {
        this.bindAuthEvents();
        this.checkAuthStatus();
    }

    /**
     * Bind authentication-related event listeners
     */
    bindAuthEvents() {
        // Auth events
        document.getElementById('loginForm').addEventListener('submit', (e) => this.handleLogin(e));
        document.getElementById('registerForm').addEventListener('submit', (e) => this.handleRegister(e));
        document.getElementById('show-register').addEventListener('click', (e) => this.showRegister(e));
        document.getElementById('show-login').addEventListener('click', (e) => this.showLogin(e));
        document.getElementById('logout-btn').addEventListener('click', (e) => this.handleLogout(e));
    }

    /**
     * Check if user has a valid session
     */
    async checkAuthStatus() {
        const sessionCookie = this.getCookie('session_id');
        
        // If there's no session cookie on the client, don't call the server —
        // go straight to the login/auth view. This avoids sending a bogus
        // header value like "null" which would force an unnecessary round
        // trip to the validate endpoint.
        if (!sessionCookie) {
            console.log('No session cookie present — redirecting to login');
            this.app.clearState();
            this.showAuth();
            return;
        }

        await fetch('/validate-session', {
            method: 'GET',
            credentials: 'include',
            headers: { 'X-Session-ID': sessionCookie }
        })
        .then(async response => {
            if (response.ok) {
                await this.app.loadDashboard();
            } else {
                this.app.clearState();
                this.showAuth();
            }
        })
        .catch(error => {
            console.error('Error validating session:', error);
            this.app.clearState();
            this.showAuth();
        });
    }

    /**
     * Show authentication UI
     */
    showAuth() {
        document.getElementById('auth-container').style.display = 'flex';
        document.getElementById('app-container').style.display = 'none';
    }

    /**
     * Show main application UI
     */
    showApp() {
        document.getElementById('auth-container').style.display = 'none';
        document.getElementById('app-container').style.display = 'block';
    }

    /**
     * Switch to registration form
     */
    showRegister(e) {
        e.preventDefault();
        document.getElementById('login-form').style.display = 'none';
        document.getElementById('register-form').style.display = 'block';
    }

    /**
     * Switch to login form
     */
    showLogin(e) {
        e.preventDefault();
        document.getElementById('login-form').style.display = 'block';
        document.getElementById('register-form').style.display = 'none';
    }

    /**
     * Handle user login
     */
    async handleLogin(e) {
        e.preventDefault();
        this.app.ui.showLoading();

        const username = document.getElementById('login-username').value;
        const password = document.getElementById('login-password').value;

        try {
            const response = await fetch('/login', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    nickname: username,
                    email: username,
                    password: password
                })
            });

            let data;
            const responseText = await response.text();
            try {
                data = JSON.parse(responseText);
            } catch (parseError) {
                console.error("Failed to parse login response:", parseError);
                data = { error: responseText };
            }

            if (response.ok) {
                // Store the user info from login response
                this.currentUser = {
                    nickname: data.user
                };

                await this.checkAuthStatus();
                this.app.ui.showToast('Login successful!', 'success');
                
            } else {
                this.app.ui.showToast(data.error || 'Login failed', 'error');
                console.error('Login error:', data);
            }
        } catch (error) {
            this.app.ui.showToast('Network error. Please try again.', 'error');
        } finally {
            this.app.ui.hideLoading();
        }
    }

    /**
     * Handle user registration
     */
    async handleRegister(e) {
        e.preventDefault();
        this.app.ui.showLoading();

        // Get the form element and create FormData directly from it
        const form = document.getElementById('registerForm');
        const formData = new FormData(form);

        try {
            const response = await fetch('/register', {
                method: 'POST',
                body: formData
            });

            let data;
            const responseText = await response.text();
            
            try {
                data = JSON.parse(responseText);
            } catch (parseError) {
                console.error('Failed to parse JSON:', responseText);
                data = { error: 'Server returned invalid response: ' + responseText };
            }

            if (response.ok) {
                this.app.ui.showToast('Registration successful! Please log in.', 'success');
                this.showLogin(e);
                form.reset();
            } else {
                this.app.ui.showToast(data.error || 'Registration failed', 'error');
                console.error('Registration error:', data);
            }
        } catch (error) {
            this.app.ui.showToast('Network error. Please try again.', 'error');
            console.error('Network error:', error);
        } finally {
            this.app.ui.hideLoading();
        }
    }

    /**
     * Handle user logout
     */
    async handleLogout(e) {
        e.preventDefault();
        this.app.ui.showLoading();

        const sessionId = this.getCookie('session_id');

        try {
            const response = await fetch('/logout', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'X-Session-ID': sessionId
                }
            });

            if (response.ok) {
                // Clear all state
                this.app.clearState();
                
                // Show auth container and ensure login form is visible
                this.showAuth();
                document.getElementById('login-form').style.display = 'block';
                document.getElementById('register-form').style.display = 'none';
                
                this.app.ui.showToast('Logout successful!', 'success');
            } else {
                const data = await response.json();
                this.app.ui.showToast(data.error || 'Logout failed', 'error');
                console.error('Logout error:', data);
            }
        } catch (error) {
            this.app.ui.showToast('Network error. Please try again.', 'error');
            console.error('Network error:', error);
        } finally {
            this.app.ui.hideLoading();
        }
    }

    /**
     * Update user display in UI
     */
    updateUserDisplay() {
        if (this.currentUser) {
            const usernameElement = document.getElementById('username-display');
            if (usernameElement) {
                usernameElement.textContent = this.currentUser.nickname;
            }
        }
    }

    /**
     * Get user information
     */
    getCurrentUser() {
        return this.currentUser;
    }

    /**
     * Set current user
     */
    setCurrentUser(user) {
        this.currentUser = user;
        this.updateUserDisplay();
    }

    /**
     * Clear current user
     */
    clearCurrentUser() {
        this.currentUser = null;
        const usernameElement = document.getElementById('username-display');
        if (usernameElement) {
            usernameElement.textContent = '';
        }
    }

    /**
     * Get cookie value by name
     */
    getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) {
            const cookieValue = parts.pop().split(';').shift();
            return cookieValue;
        }
        return null;
    }
}