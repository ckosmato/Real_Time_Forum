class ForumApp {
    constructor() {
        this.currentUser = null;
        this.currentView = 'home';
        this.categories = [];
        this.posts = [];
        this.selectedPostId = null;
        this.websocket = null;
        this.isWebSocketConnected = false;
        this.currentChatUser = null;
        
        this.init();
    }

    init() {
        this.bindEvents();
        this.checkAuthStatus();
        this.initializeWidget();
    }

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

    bindEvents() {
        // Auth events
        document.getElementById('loginForm').addEventListener('submit', (e) => this.handleLogin(e));
        document.getElementById('registerForm').addEventListener('submit', (e) => this.handleRegister(e));
        document.getElementById('show-register').addEventListener('click', (e) => this.showRegister(e));
        document.getElementById('show-login').addEventListener('click', (e) => this.showLogin(e));
        document.getElementById('logout-btn').addEventListener('click', (e) => this.handleLogout(e));

        // Navigation events
        document.getElementById('home-btn').addEventListener('click', () => this.showView('home'));
        document.getElementById('my-posts-btn').addEventListener('click', () => this.showView('my-posts'));
        document.getElementById('create-post-btn').addEventListener('click', () => this.showView('create-post'));

        // Post creation
        document.getElementById('createPostForm').addEventListener('submit', (e) => this.handleCreatePost(e));
    }

    async checkAuthStatus() {
       
        const sessionCookie = this.getCookie('session_id');
       
        if (sessionCookie) {     
            await this.loadDashboard();
        } else {
            this.clearState();
            this.showAuth();
            document.getElementById('login-form').style.display = 'block';
            document.getElementById('register-form').style.display = 'none';
            console.log("not logged in");
            // Still render categories even when not logged in
            //this.renderCategories();
        }
    }

    showAuth() {
        document.getElementById('auth-container').style.display = 'flex';
        document.getElementById('app-container').style.display = 'none';
    }

    showApp() {
        document.getElementById('auth-container').style.display = 'none';
        document.getElementById('app-container').style.display = 'block';
    }

    showRegister(e) {
        e.preventDefault();
        document.getElementById('login-form').style.display = 'none';
        document.getElementById('register-form').style.display = 'block';
    }

    showLogin(e) {
        e.preventDefault();
        document.getElementById('login-form').style.display = 'block';
        document.getElementById('register-form').style.display = 'none';
    }

    async handleLogin(e) {
        e.preventDefault();
        this.showLoading();

        const username = document.getElementById('login-username').value;
        const password = document.getElementById('login-password').value;

        try {
            const response = await fetch('/login', {
                method: 'POST',
                credentials: 'include', // Changed from 'same-origin' to 'include' to ensure cookies are sent/received
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
                this.showToast('Login successful!', 'success');
                
            } else {
                this.showToast(data.error || 'Login failed', 'error');
                console.error('Login error:', data);
            }
        } catch (error) {
            this.showToast('Network error. Please try again.', 'error');
        } finally {
            this.hideLoading();
        }
    }

    async handleRegister(e) {
        e.preventDefault();
        this.showLoading();

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
                this.showToast('Registration successful! Please log in.', 'success');
                this.showLogin(e);
                form.reset();
            } else {
                this.showToast(data.error || 'Registration failed', 'error');
                console.error('Registration error:', data);
            }
        } catch (error) {
            this.showToast('Network error. Please try again.', 'error');
            console.error('Network error:', error);
        } finally {
            this.hideLoading();
        }
    }

    async handleLogout(e) {
        e.preventDefault();
        this.showLoading();

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
                this.clearState();
                
                // Show auth container and ensure login form is visible
                this.showAuth();
                document.getElementById('login-form').style.display = 'block';
                document.getElementById('register-form').style.display = 'none';
                
                this.showToast('Logout successful!', 'success');
            } else {
                const data = await response.json();
                this.showToast(data.error || 'Logout failed', 'error');
                console.error('Logout error:', data);
            }
        } catch (error) {
            this.showToast('Network error. Please try again.', 'error');
            console.error('Network error:', error);
        } finally {
            this.hideLoading();
        }
    }

    async loadDashboard() {

        this.showLoading();
        //get session id from cookie
        const sessionId = this.getCookie('session_id');


        try {
            // Show the dashboard UI immediately after successful login
            this.showApp();
            this.showView('home');
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
            console.log('Dashboard response status:', response.status);
            
            const data = await response.json();
            console.log('Dashboard response data:', data);
            
            if (response.ok) {
                // Update user information from the server response
                if (data.user) {
                    this.currentUser = {
                        nickname: data.user.Nickname,
                    };
                    this.updateUserDisplay();
                } else {
                    this.clearState();
                    this.showAuth();
                    document.getElementById('login-form').style.display = 'block';
                    document.getElementById('register-form').style.display = 'none';
                    console.error('No user data in dashboard response');
                }

                this.categories = data.categories || [];
                this.posts = data.posts || [];
                
                console.log('Loaded categories:', this.categories.length);
                console.log('Loaded posts:', this.posts.length);
                
                this.renderCategories();
                this.renderPosts();
            } else {
                console.error('Failed to load posts:', data.error);
                this.showToast('Failed to load posts. Please refresh the page.', 'error');
                // Still render categories even if the request failed
                this.renderCategories();
            }
        } catch (error) {
            console.error('Dashboard error:', error);
            this.showToast('Failed to load content. Please refresh the page.', 'error');
            // Still render categories even if there's a network error
            this.renderCategories();
        } finally {
            this.hideLoading();
            // Initialize chat event listeners after dashboard is loaded
            this.initializeChatEventListeners();
            // Connect WebSocket for real-time features
            this.connectWebSocket();
            // Request notification permission
            this.requestNotificationPermission();
        }
    }

    updateUserDisplay() {
        if (this.currentUser) {
            const usernameElement = document.getElementById('username-display');
            if (usernameElement) {
                usernameElement.textContent = this.currentUser.nickname;
            }
        }
    }

    // WebSocket Connection Methods
    connectWebSocket() {
        if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
            console.log('WebSocket already connected');
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        console.log('Connecting to WebSocket:', wsUrl);
        
        this.websocket = new WebSocket(wsUrl);
        
        this.websocket.onopen = () => {
            console.log('WebSocket connected');
            this.isWebSocketConnected = true;
        };
        
        this.websocket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleWebSocketMessage(message);
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        };
        
        this.websocket.onclose = (event) => {
            console.log('WebSocket disconnected:', event.code, event.reason);
            this.isWebSocketConnected = false;
            
            // Attempt to reconnect after 3 seconds if not intentional
            if (event.code !== 1000) {
                setTimeout(() => {
                    console.log('Attempting to reconnect WebSocket...');
                    this.connectWebSocket();
                }, 3000);
            }
        };
        
        this.websocket.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.isWebSocketConnected = false;
        };
    }

    disconnectWebSocket() {
        if (this.websocket) {
            this.websocket.close(1000, 'User logged out');
            this.websocket = null;
            this.isWebSocketConnected = false;
        }
    }

    handleWebSocketMessage(message) {
        console.log('Received WebSocket message:', message);
        console.log('Current user:', this.currentUser?.nickname);
        
        switch (message.type) {
            case 'chat_message':
                console.log('Processing chat message...');
                this.displayChatMessage(message);
                break;
            case 'user_online':
                this.handleUserOnline(message);
                break;
            case 'user_offline':
                this.handleUserOffline(message);
                break;
            default:
                console.log('Unknown message type:', message.type);
        }
    }

    sendWebSocketMessage(message) {
        if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
            this.websocket.send(JSON.stringify(message));
            return true;
        } else {
            console.error('WebSocket not connected');
            this.showToast('Connection lost. Please refresh the page.', 'error');
            return false;
        }
    }

    displayChatMessage(message) {
        console.log('displayChatMessage called with:', message);
        
        const chatMessages = document.getElementById('chat-messages');
        const chatWidget = document.getElementById('chat-widget');
        
        if (!chatMessages) {
            console.error('Chat messages element not found');
            return;
        }
        
        const isFromCurrentUser = message.from === this.currentUser?.nickname;
        const isToCurrentUser = message.to === this.currentUser?.nickname;
        
        console.log('Message analysis:', {
            from: message.from,
            to: message.to,
            currentUser: this.currentUser?.nickname,
            isFromCurrentUser,
            isToCurrentUser
        });
        
        // If this is an incoming message (not from current user but to current user)
        if (!isFromCurrentUser && isToCurrentUser) {
            console.log('Processing incoming message...');
            
            // Show notification
            this.showChatNotification(message);
            
            // Auto-open chat if it's not open
            if (!chatWidget || chatWidget.style.display === 'none') {
                console.log('Auto-opening chat for new message');
                this.autoOpenChatForNewMessage(message);
                return; // Let the auto-open handle message display
            }
        }
        
        // Only display in chat if chat widget exists and is visible
        if (!chatWidget || chatWidget.style.display === 'none') {
            console.log('Chat widget not visible, skipping message display');
            return;
        }
        
        // Check if this message is for the current open chat
        if (this.currentChatUser && 
            (message.from === this.currentChatUser || message.to === this.currentChatUser ||
             (isFromCurrentUser && message.to === this.currentChatUser))) {
            
            console.log('Adding message to current chat');
            
            const messageDiv = document.createElement('div');
            
            messageDiv.className = `chat-message ${isFromCurrentUser ? 'sent' : 'received'}`;
            messageDiv.innerHTML = `
                ${this.escapeHtml(message.content)}
                <div class="chat-message-time">${new Date(message.timestamp).toLocaleTimeString()}</div>
            `;
            
            chatMessages.appendChild(messageDiv);
            chatMessages.scrollTop = chatMessages.scrollHeight;
        } else {
            console.log('Message not for current chat user');
        }
    }

    showChatNotification(message) {
        // Show browser notification if permission granted
        if (Notification.permission === 'granted') {
            const notification = new Notification(`New message from ${message.from}`, {
                body: message.content,
                icon: '/favicon.ico', // You can add a favicon
                tag: `chat-${message.from}`, // Replace previous notifications from same user
            });
            
            // Auto close after 5 seconds
            setTimeout(() => {
                notification.close();
            }, 5000);
            
            // Click notification to open chat
            notification.onclick = () => {
                this.openChatWithUser(message.from);
                window.focus();
                notification.close();
            };
        }
        
        // Show in-app notification toast
        this.showToast(`💬 New message from ${message.from}`, 'info', 4000);
        
        // Add visual indicator to the sender in online users list
        this.addMessageIndicatorToUser(message.from);
    }

    autoOpenChatForNewMessage(message) {
        console.log('Auto-opening chat for new message from:', message.from);
        
        // Open chat with the sender
        this.openChatWithUser(message.from);
        
        // Add the message to the chat
        const chatMessages = document.getElementById('chat-messages');
        if (chatMessages) {
            // Clear the default welcome message
            chatMessages.innerHTML = '';
            
            // Add the received message
            const messageDiv = document.createElement('div');
            messageDiv.className = 'chat-message received';
            messageDiv.innerHTML = `
                ${this.escapeHtml(message.content)}
                <div class="chat-message-time">${new Date(message.timestamp).toLocaleTimeString()}</div>
            `;
            
            chatMessages.appendChild(messageDiv);
            chatMessages.scrollTop = chatMessages.scrollHeight;
        }
        
        // Show a notification that chat was auto-opened
        this.showToast(`📱 Chat opened with ${message.from}`, 'success', 3000);
    }

    addMessageIndicatorToUser(username) {
        // Find the user in the online users list
        const activeUsersList = document.getElementById('active-users-list');
        if (!activeUsersList) return;
        
        const userElements = activeUsersList.querySelectorAll('.active-user');
        userElements.forEach(userElement => {
            if (userElement.textContent.trim() === username) {
                // Add a message indicator if it doesn't exist
                let indicator = userElement.querySelector('.message-indicator');
                if (!indicator) {
                    indicator = document.createElement('span');
                    indicator.className = 'message-indicator';
                    indicator.textContent = '1';
                    userElement.appendChild(indicator);
                } else {
                    // Increment the count
                    const count = parseInt(indicator.textContent) || 0;
                    indicator.textContent = (count + 1).toString();
                }
            }
        });
    }

    removeMessageIndicatorFromUser(username) {
        // Find the user in the online users list and remove indicator
        const activeUsersList = document.getElementById('active-users-list');
        if (!activeUsersList) return;
        
        const userElements = activeUsersList.querySelectorAll('.active-user');
        userElements.forEach(userElement => {
            if (userElement.textContent.replace(/\d+$/, '').trim() === username) {
                const indicator = userElement.querySelector('.message-indicator');
                if (indicator) {
                    indicator.remove();
                }
            }
        });
    }

    requestNotificationPermission() {
        if ('Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission().then(permission => {
                console.log('Notification permission:', permission);
            });
        }
    }

    handleUserOnline(message) {
        console.log('User came online:', message.user);
        // Refresh the active users list
        this.loadActiveUsers();
    }

    handleUserOffline(message) {
        console.log('User went offline:', message.user);
        // Refresh the active users list
        this.loadActiveUsers();
    }

    handleUserOnline(message) {
        console.log(`User ${message.username} came online`);
        // Refresh the active users list
        this.loadActiveUsers();
    }

    handleUserOffline(message) {
        console.log(`User ${message.username} went offline`);
        // Refresh the active users list
        this.loadActiveUsers();
    }

    async loadActiveUsers() {
        this.showLoading();

        try {
            const response = await fetch('/dashboard/active-users', {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'X-Session-ID': this.getCookie('session_id')
                }
               
            });

            if (response.ok) {
                const data = await response.json();
                this.renderActiveUsers(data.users);
            } else {
                console.error('Failed to load active users');
            }
        } catch (error) {
            console.error('Error loading active users:', error);
        } finally {
            this.hideLoading();
        }
    }

    renderActiveUsers(users) {
        const container = document.getElementById('active-users-list');
        container.innerHTML = '';
        users.forEach(user => {
            const userDiv = document.createElement('div');
            userDiv.className = 'active-user';
            userDiv.textContent = this.escapeHtml(user.Nickname);
            userDiv.dataset.username = user.Nickname;
            userDiv.addEventListener('click', (e) => {
                console.log('User clicked:', user.Nickname); // Debug log
                e.stopPropagation(); // Prevent widget toggle when clicking user
                this.openChatWithUser(user.Nickname);
            });
            container.appendChild(userDiv);
        });
    }

    // Toggle the online users widget open/closed
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

    openChatWithUser(username) {
        console.log('Opening chat with user:', username); // Debug log
        
        const chatWidget = document.getElementById('chat-widget');
        const chatUsername = document.getElementById('chat-username');
        const chatMessages = document.getElementById('chat-messages');
        
        console.log('Chat elements found:', {
            chatWidget: !!chatWidget,
            chatUsername: !!chatUsername,
            chatMessages: !!chatMessages
        }); // Debug log
        
        if (chatWidget && chatUsername && chatMessages) {
            // Set the current chat user
            this.currentChatUser = username;
            
            // Remove message indicator for this user
            this.removeMessageIndicatorFromUser(username);
            
            // Set the chat user in UI
            chatUsername.innerHTML = `<i class="fa-solid fa-comment"></i> Chat with ${this.escapeHtml(username)}`;
            
            // Clear previous messages 
            chatMessages.innerHTML = '<div class="chat-message received">Start your conversation with ' + this.escapeHtml(username) + '!</div>';
            
            // Show the chat widget
            chatWidget.style.display = 'flex';
            
            // Focus on the input
            const chatInput = document.getElementById('chat-input');
            if (chatInput) chatInput.focus();
            
            console.log('Chat widget should now be visible'); // Debug log
        } else {
            console.error('Chat elements not found!');
        }
    }

    closeChatWidget() {
        const chatWidget = document.getElementById('chat-widget');
        if (chatWidget) {
            chatWidget.style.display = 'none';
            this.currentChatUser = null; // Clear current chat user
        }
    }

    initializeChatEventListeners() {
        console.log('Initializing chat event listeners...'); // Debug log
        
        // Close chat button
        const closeBtn = document.querySelector('.chat-close');
        console.log('Close button found:', !!closeBtn); // Debug log
        if (closeBtn) {
            closeBtn.addEventListener('click', () => this.closeChatWidget());
        }

        // Send message functionality
        const sendBtn = document.getElementById('chat-send');
        const chatInput = document.getElementById('chat-input');
        
        console.log('Send button and input found:', !!sendBtn, !!chatInput); // Debug log
        
        if (sendBtn && chatInput) {
            sendBtn.addEventListener('click', () => this.sendChatMessage());
            chatInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.sendChatMessage();
                }
            });
        }
        
        console.log('Chat event listeners initialized'); // Debug log
    }

    // Test function to manually open chat (for debugging)
    testOpenChat() {
        console.log('Test: Opening chat with test user...');
        this.openChatWithUser('TestUser');
    }

    sendChatMessage() {
        const chatInput = document.getElementById('chat-input');
        const chatMessages = document.getElementById('chat-messages');
        
        if (chatInput && chatMessages && chatInput.value.trim() && this.currentChatUser) {
            const messageContent = chatInput.value.trim();
            
            // Create WebSocket message
            const message = {
                type: 'chat_message',
                to: this.currentChatUser,
                content: messageContent,
                timestamp: new Date().toISOString()
            };
            
            // Send via WebSocket
            if (this.sendWebSocketMessage(message)) {
                // Only add to UI locally if WebSocket send was successful
                // The actual message will be received back through WebSocket
                
                // Clear input
                chatInput.value = '';
                
                console.log('Message sent via WebSocket:', message);
            } else {
                // Fallback: show error message
                this.showToast('Failed to send message. Please check your connection.', 'error');
            }
        } else if (!this.currentChatUser) {
            console.error('No chat user selected');
        }
    }

    renderCategories() {
        const container = document.getElementById('categories-list');
        
        // Check if categories are missing and restore them if needed
        const categoryItems = container.querySelectorAll('.category-item');
        if (categoryItems.length === 0) {
            // Restore hardcoded categories if they're missing
            container.innerHTML = `
                <div class="category-item active">All Posts</div>
                <div class="category-item">General Discussion</div>
                <div class="category-item">Sports</div>
                <div class="category-item">Music</div>
                <div class="category-item">Movies & TV</div>
                <div class="category-item">Books</div>
                <div class="category-item">Science</div>
                <div class="category-item">News</div>
            `;
        }

        // Add click handlers to categories
        const updatedCategoryItems = container.querySelectorAll('.category-item');
        updatedCategoryItems.forEach((item, index) => {
            // Remove existing click listeners to avoid duplicates
            item.replaceWith(item.cloneNode(true));
        });

        // Re-select after cloning (to get fresh elements without old listeners)
        const finalCategoryItems = container.querySelectorAll('.category-item');
        finalCategoryItems.forEach((item, index) => {
            if (index === 0) {
                // First item is "All Posts"
                item.addEventListener('click', () => this.filterByCategory(null));
            } else {
                // For hardcoded categories, use index as category ID
                item.addEventListener('click', () => this.filterByCategory(index));
            }
        });

        // If we have categories from backend, we could append them here
        // but for now, just use the hardcoded ones
    }


    async filterByCategory(categoryId) {
        // Update active category
        document.querySelectorAll('.category-item').forEach(item => {
            item.classList.remove('active');
        });
        event.target.classList.add('active');

        this.showLoading();

        try {
            let url = '/dashboard';
            if (categoryId) {
                url = `/category/${categoryId}`;
            }

            // Get session id from cookie
            const sessionId = this.getCookie('session_id');

            const response = await fetch(url, {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'X-Session-ID': sessionId
                }
            });
            const data = await response.json();

            if (response.ok) {
                this.posts = data.posts || [];
                this.renderPosts();
            } else {
                this.showToast('Failed to load posts', 'error');
            }
        } catch (error) {
            this.showToast('Network error', 'error');
        } finally {
            this.hideLoading();
        }
    }

    renderPosts() {
        const container = document.getElementById('posts-container');
        container.innerHTML = '';

        if (this.posts.length === 0) {
            container.innerHTML = '<p style="text-align: center; color: #999;">No posts found.</p>';
            return;
        }

        this.posts.forEach(post => {
            const postCard = this.createPostCard(post);
            container.appendChild(postCard);
        });
    }

    createPostCard(post) {
        const card = document.createElement('div');
        card.className = 'post-card';
        card.addEventListener('click', () => this.viewPost(post.ID)); // Changed from post.id to post.ID

        const categoriesHtml = post.Categories ? 
            post.Categories.map(cat => `<span class="category-tag">${cat}</span>`).join('') : '';

        card.innerHTML = `
            <div class="post-header">
                <h3 class="post-title">${this.escapeHtml(post.Title)}</h3>
                <div class="post-meta">
                    <div>By: ${this.escapeHtml(post.AuthorName)}</div>
                    <div>${this.formatDate(post.CreatedAt)}</div>
                </div>
            </div>
            <div class="post-content">
                ${this.escapeHtml((post.Content || '').substring(0, 200))}${(post.Content || '').length > 200 ? '...' : ''}
            </div>
            <div class="post-categories">
                ${categoriesHtml}
            </div>
        `;

        return card;
    }

    async viewPost(postId) {
        console.log('ViewPost called with postId:', postId, 'type:', typeof postId);
        this.selectedPostId = postId;
        this.showLoading();

        try {
            const url = `/post?id=${postId}`;
            console.log('Fetching URL:', url);
            const response = await fetch(url, { 
                credentials: 'same-origin',
                headers: {
                    'X-Session-ID': this.getCookie('session_id'),
                    'Accept': 'application/json'
                }
            });
            
            console.log('ViewPost response status:', response.status);
            
            if (response.ok) {
                const data = await response.json();
                this.renderPostDetail(data.post, data.comments || []);
                this.showView('post');
            } else {
                const data = await response.json();
                console.error('ViewPost error:', data);
                this.showToast('Failed to load post: ' + (data.error || 'Unknown error'), 'error');
            }
        } catch (error) {
            console.error('ViewPost network error:', error);
            this.showToast('Network error while loading post', 'error');
        } finally {
            this.hideLoading();
        }
    }

    renderPostDetail(post, comments) {
        const container = document.getElementById('post-detail');
        
        const categoriesHtml = post.Categories ? 
            post.Categories.map(cat => `<span class="category-tag">${cat}</span>`).join('') : '';

        const commentsHtml = comments.map(comment => `
            <div class="comment">
                <div class="comment-header">
                    <span class="comment-author">${this.escapeHtml(comment.AuthorName || comment.authorName)}</span>
                    <span class="comment-date">${this.formatDate(comment.CreatedAt || comment.createdAt)}</span>
                </div>
                <div class="comment-content">${this.escapeHtml(comment.Content || comment.content)}</div>
            </div>
        `).join('');

        container.innerHTML = `
            <div class="post-navigation">
                <button onclick="app.showView('home')" class="back-btn">← Back to Posts</button>
            </div>
            <div class="post-header">
                <h1 class="post-title">${this.escapeHtml(post.Title)}</h1>
                <div class="post-meta">
                    <div>By: ${this.escapeHtml(post.AuthorName)}</div>
                    <div>${this.formatDate(post.CreatedAt)}</div>
                </div>
            </div>
            <div class="post-content" style="margin: 2rem 0;">
                ${this.escapeHtml(post.Content || '').replace(/\n/g, '<br>')}
            </div>
            <div class="post-categories">
                ${categoriesHtml}
            </div>
            
            <div class="comments-section">
                <h3>Comments (${comments.length})</h3>
                
                ${this.currentUser ? `
                    <form class="comment-form" onsubmit="app.handleCreateComment(event)">
                        <textarea placeholder="Write your comment..." required></textarea>
                        <button type="submit">Post Comment</button>
                    </form>
                ` : '<p style="color: #999;">Please log in to comment.</p>'}
                
                <div class="comments-list">
                    ${commentsHtml || '<p style="color: #999;">No comments yet.</p>'}
                </div>
            </div>
        `;
    }

    async handleCreateComment(e) {
        e.preventDefault();
        
        const textarea = e.target.querySelector('textarea');
        const content = textarea.value.trim();
        
        if (!content) return;

        this.showLoading();

        try {
            const formData = new FormData();
            formData.append('comment', content);
            formData.append('post_id', this.selectedPostId);

            const response = await fetch('/post/createcomment', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'X-Session-ID': this.getCookie('session_id')
                },
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                this.showToast('Comment posted successfully!', 'success');
                textarea.value = '';
                // Reload post to show new comment
                await this.viewPost(this.selectedPostId);
            } else {
                this.showToast(data.error || 'Failed to post comment', 'error');
            }
        } catch (error) {
            this.showToast('Network error', 'error');
        } finally {
            this.hideLoading();
        }
    }

    async handleCreatePost(e) {
        e.preventDefault();
        this.showLoading();

        const title = document.getElementById('post-title').value;
        const content = document.getElementById('post-content').value;
        const selectedCategories = Array.from(document.querySelectorAll('#post-categories input:checked'))
            .map(cb => cb.value);

        console.log('Creating post with:', { title, content, selectedCategories });

        if (selectedCategories.length === 0) {
            this.showToast('Please select at least one category', 'error');
            this.hideLoading();
            return;
        }

        try {
            const formData = new FormData();
            formData.append('title', title);
            formData.append('content', content);
            selectedCategories.forEach(categoryId => {
                formData.append('categories', categoryId);
            });

            console.log('Sending request to /createpost...');

            const response = await fetch('/createpost', {
                method: 'POST',
                credentials: 'same-origin',
                headers: {
                    'X-Session-ID': this.getCookie('session_id')
                },
                body: formData
            });

            console.log('Response status:', response.status, 'Response ok:', response.ok);

            const data = await response.json();
            console.log('Response data:', data);

            if (response.ok) {
                this.showToast('Post created successfully!', 'success');
                document.getElementById('createPostForm').reset();
                await this.loadDashboard(); // Refresh dashboard
                this.showView('home');
            } else {
                this.showToast(data.error || 'Failed to create post', 'error');
            }
        } catch (error) {
            console.error('Network error details:', error);
            this.showToast('Network error', 'error');
        } finally {
            this.hideLoading();
        }
    }

    async loadMyPosts() {
        if (!this.currentUser) return;
        
        this.showLoading();

        try {
            // This would need a new endpoint in your backend
            const response = await fetch(`/dashboard/my-posts`, {
                method: 'GET',
                headers: {
                    'X-SESSION-ID': this.getCookie('session_id')
                },
                credentials: 'same-origin'
            });

            if (response.ok) {
                const data = await response.json();
                const container = document.getElementById('my-posts-container');
                container.innerHTML = '';
                
                if (data.posts && data.posts.length > 0) {
                    data.posts.forEach(post => {
                        const postCard = this.createPostCard(post);
                        container.appendChild(postCard);
                    });
                } else {
                    container.innerHTML = '<p style="text-align: center; color: #999;">You haven\'t created any posts yet.</p>';
                }
            } else {
                this.showToast('Failed to load your posts', 'error');
            }
        } catch (error) {
            this.showToast('Network error', 'error');
        } finally {
            this.hideLoading();
        }
    }

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
            this.loadMyPosts();
        }
    }

    showLoading() {
        document.getElementById('loading').style.display = 'flex';
    }

    hideLoading() {
        document.getElementById('loading').style.display = 'none';
    }

    showToast(message, type = 'info', timeout = 4000) {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;

        document.getElementById('toast-container').appendChild(toast);

        setTimeout(() => {
            toast.remove();
        }, timeout);
    }

    getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) {
            const cookieValue = parts.pop().split(';').shift();
            return cookieValue;
        }
        return null;
    }

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

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
    }
    clearState() {
        // Clear all state
        this.currentUser = null;
        this.categories = [];
        this.posts = [];
        this.selectedPostId = null;
        this.currentChatUser = null;
        
        // Disconnect WebSocket
        this.disconnectWebSocket();
        
        // Clear any displayed content
        document.getElementById('posts-container').innerHTML = '';
        document.getElementById('my-posts-container').innerHTML = '';
        
        // Reset username display
        document.getElementById('username-display').textContent = '';
        
        // Close chat widget
        this.closeChatWidget();
    }
}



// Initialize the app
const app = new ForumApp();

// Make handleCreateComment available globally for the onclick handler
window.app = app;