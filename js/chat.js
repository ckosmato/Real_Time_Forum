 /**
 * WebSocket and Chat Module
 * Handles real-time messaging, WebSocket connections, and chat functionality
 */
class ChatManager {
    constructor(app) {
        this.app = app;
        this.websocket = null;
        this.isWebSocketConnected = false;
        this.currentChatUser = null;
        this.onlineUsers = []; // Store online users list
        
        // Pagination properties
        this.currentOffset = 0;
        this.messagesPerPage = 10;
        this.isLoadingHistory = false;
        this.hasMoreMessages = true;
        this.scrollThrottleTimer = null;
    }

    /**
     * Initialize chat functionality
     */
    init() {
        this.initializeChatEventListeners();
        this.connectWebSocket();
    }

    /**
     * Initialize chat-related event listeners
     */
    initializeChatEventListeners() {
        // Close chat button
        const closeBtn = document.querySelector('.chat-close');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => this.closeChatWidget());
        }

        // Send message functionality
        const sendBtn = document.getElementById('chat-send');
        const chatInput = document.getElementById('chat-input');
        
        if (sendBtn && chatInput) {
            sendBtn.addEventListener('click', () => this.sendChatMessage());
            chatInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.sendChatMessage();
                }
            });
        }
    }

    /**
     * Connect to WebSocket server
     */
    connectWebSocket() {
        if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
            return;
        }

        const sessionId = this.app.auth.getCookie('session_id');

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?session_id=${sessionId}`;

        this.websocket = new WebSocket(wsUrl);

        this.websocket.onopen = () => {
            console.log('WebSocket connected');
            this.isWebSocketConnected = true;
        };

        this.websocket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                
                // Handle different message types
                switch (message.type) {
                    case 'chat_message':
                        this.displayChatMessage(message);
                        break;
                    case 'user_joined':
                        this.handleUserJoined(message);
                        break;
                    case 'user_left':
                        this.handleUserLeft(message);
                        break;
                    case 'initial_online_users':
                        this.handleInitialOnlineUsers(message);
                        break;
                    case 'online_users_update':
                        this.handleOnlineUsersUpdate(message);
                        break;
                    default:
                        // Default to chat message for backward compatibility
                        this.displayChatMessage(message);
                        break;
                }
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        };

        this.websocket.onclose = (event) => {
            console.log('WebSocket disconnected:', event.code, event.reason);
            this.isWebSocketConnected = false;

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

    /**
     * Disconnect WebSocket
     */
    disconnectWebSocket() {
        if (this.websocket) {
            this.websocket.close(1000, 'User logged out');
            this.websocket = null;
            this.isWebSocketConnected = false;
        }
    }

    /**
     * Send message via WebSocket
     */
    sendWebSocketMessage(message) {
        if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
            this.websocket.send(JSON.stringify(message));
            return true;
        } else {
            console.error('WebSocket not connected');
            this.app.ui.showToast('Connection lost. Please refresh the page.', 'error');
            return false;
        }
    }

    /**
     * Display chat message in UI
     */
    displayChatMessage(message) {
        const chatMessages = document.getElementById('chat-messages');
        const chatWidget = document.getElementById('chat-widget');
        
        if (!chatMessages) {
            console.error('Chat messages element not found');
            return;
        }
        
        const currentUser = this.app.auth.getCurrentUser();
        const isFromCurrentUser = message.from === currentUser?.nickname;
        const isToCurrentUser = message.to === currentUser?.nickname;
        
        // If this is an incoming message (not from current user but to current user)
        if (!isFromCurrentUser && isToCurrentUser) {
            // Show notification
            this.app.ui.showToast(`💬 New message from ${message.from}`, 'info', 4000);
            
            // Auto-open chat if it's not open
            if (!chatWidget || chatWidget.style.display === 'none') {
                this.autoOpenChatForNewMessage(message);
                return; // Let the auto-open handle message display
            }
        }
        
        // Only display in chat if chat widget exists and is visible
        if (!chatWidget || chatWidget.style.display === 'none') {
            return;
        }
        
        // Check if this message is for the current open chat
        if (this.currentChatUser && 
            (message.from === this.currentChatUser || message.to === this.currentChatUser ||
             (isFromCurrentUser && message.to === this.currentChatUser))) {
            
            const messageDiv = document.createElement('div');
            
            messageDiv.className = `chat-message ${isFromCurrentUser ? 'sent' : 'received'}`;
            messageDiv.innerHTML = `
                ${this.app.ui.escapeHtml(message.content)}
                <div class="chat-message-time">${new Date(message.timestamp).toLocaleTimeString()}</div>
            `;
            
            chatMessages.appendChild(messageDiv);
            chatMessages.scrollTop = chatMessages.scrollHeight;
        }
    }

    /**
     * Auto-open chat for new incoming message
     */
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
                ${this.app.ui.escapeHtml(message.content)}
                <div class="chat-message-time">${new Date(message.timestamp).toLocaleTimeString()}</div>
            `;
            
            chatMessages.appendChild(messageDiv);
            chatMessages.scrollTop = chatMessages.scrollHeight;
        }
    }

    /**
     * Open chat with specific user
     */
    openChatWithUser(username) {
        const chatWidget = document.getElementById('chat-widget');
        const chatUsername = document.getElementById('chat-username');
        const chatMessages = document.getElementById('chat-messages');
        
        if (chatWidget && chatUsername && chatMessages) {
            // Set the current chat user
            this.currentChatUser = username;
            
            // Reset pagination state
            this.currentOffset = 0;
            this.hasMoreMessages = true;
            this.isLoadingHistory = false;
            
            // Set the chat user in UI
            chatUsername.innerHTML = `<i class="fa-solid fa-comment"></i> Chat with ${this.app.ui.escapeHtml(username)}`;
            
            // Show the chat widget
            chatWidget.style.display = 'flex';
            
            // Focus on the input
            const chatInput = document.getElementById('chat-input');
            if (chatInput) {
                chatInput.focus();
            }
            
            // Add scroll event listener for pagination
            this.setupScrollPagination(chatMessages);
            
            // Load initial batch of messages (20 for initial view)
            this.loadInitialChatHistory(username);
        }
    }

    /**
     * Close chat widget
     */
    closeChatWidget() {
        const chatWidget = document.getElementById('chat-widget');
        if (chatWidget) {
            chatWidget.style.display = 'none';
            this.currentChatUser = null; // Clear current chat user
            
            // Reset pagination state
            this.currentOffset = 0;
            this.hasMoreMessages = true;
            this.isLoadingHistory = false;
            
            // Remove scroll event listener
            const chatMessages = document.getElementById('chat-messages');
            if (chatMessages) {
                chatMessages.removeEventListener('scroll', this.throttledScrollHandler);
            }
        }
    }

    /**
     * Setup scroll-based pagination for chat messages
     */
    setupScrollPagination(chatMessages) {
        // Remove any existing scroll listener
        if (this.throttledScrollHandler) {
            chatMessages.removeEventListener('scroll', this.throttledScrollHandler);
        }
        
        // Create throttled scroll handler
        this.throttledScrollHandler = this.throttle((e) => {
            const element = e.target;
            
            // Check if user scrolled to the top (with some threshold)
            if (element.scrollTop <= 50 && this.hasMoreMessages && !this.isLoadingHistory) {
                console.log('Loading more messages...');
                this.loadMoreMessages();
            }
        }, 300); // Throttle to 300ms
        
        // Add scroll event listener
        chatMessages.addEventListener('scroll', this.throttledScrollHandler);
    }

    /**
     * Throttle function to prevent spam scroll events
     */
    throttle(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func.apply(this, args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    /**
     * Send chat message
     */
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
                chatInput.value = '';
            } else {
                // Fallback: show error message
                this.app.ui.showToast('Failed to send message. Please check your connection.', 'error');
            }
        } else if (!this.currentChatUser) {
            console.error('No chat user selected');
        }
    }

    /**
     * Load initial chat history with a larger batch for better UX
     */
    loadInitialChatHistory(username) {
        console.log("Loading initial chat history for", username);
        const chatMessages = document.getElementById('chat-messages');

        this.isLoadingHistory = true;

        // Load only 10 messages initially (last 10 messages)
        const initialLimit = 10;

        fetch(`/chathistory?user2=${encodeURIComponent(username)}&limit=${initialLimit}&offset=0`, {
            method: 'GET',
            headers: {
                'X-Session-ID': this.app.auth.getCookie('session_id'),
            },
            credentials: 'include'
        })
        .then(response => {
            console.log("Response status:", response.status);
            return response.json();
        })
        .then(data => {
            const currentUser = this.app.auth.getCurrentUser();
            
            console.log(`Initial load: received ${data.history?.length || 0} messages`);
            console.log('Has more messages:', data.hasMore);
            
            // Clear and populate with initial messages
            chatMessages.innerHTML = '';
            
            // Check if there are more messages beyond our initial load
            this.hasMoreMessages = data.hasMore || false;
            
            if (data.history && data.history.length > 0) {
                data.history.forEach((msg, index) => {
                    console.log(`Message ${index + 1}: "${msg.content}" at ${msg.timestamp}`);
                    const msgDiv = document.createElement('div');
                    msgDiv.classList.add('chat-message');

                    // Check if the message is from the current user or the other user
                    if (msg.from === currentUser?.nickname) {
                        msgDiv.classList.add('sent');       // message you sent
                    } else {
                        msgDiv.classList.add('received');   // message received
                    }

                    // Add the content
                    msgDiv.textContent = msg.content;

                    // Add timestamp
                    const timeDiv = document.createElement('div');
                    timeDiv.classList.add('chat-message-time');
                    timeDiv.textContent = new Date(msg.timestamp).toLocaleTimeString();
                    msgDiv.appendChild(timeDiv);

                    chatMessages.appendChild(msgDiv);
                });
                
                // Set offset to the number of messages we loaded
                this.currentOffset = data.history.length;
                
                // Scroll to bottom to show most recent messages
                chatMessages.scrollTop = chatMessages.scrollHeight;
            }
            
            this.isLoadingHistory = false;
        })
        .catch(err => {
            console.error('Error fetching initial chat history:', err);
            chatMessages.innerHTML = '<div class="chat-message error">Failed to load chat history.</div>';
            this.isLoadingHistory = false;
        });
    }

    /**
     * Load chat history with a user (for pagination)
     */
    /**
     * Load chat history with a user (for pagination)
     */
    loadChatHistory(username) {
        console.log("Loading more chat history for", username, "offset:", this.currentOffset);
        const chatMessages = document.getElementById('chat-messages');

        this.isLoadingHistory = true;

        fetch(`/chathistory?user2=${encodeURIComponent(username)}&limit=${this.messagesPerPage}&offset=${this.currentOffset}`, {
            method: 'GET',
            headers: {
                'X-Session-ID': this.app.auth.getCookie('session_id'),
            },
            credentials: 'include'
        })
        .then(response => {
            console.log("Response status:", response.status);
            return response.json();
        })
        .then(data => {
            const currentUser = this.app.auth.getCurrentUser();
            
            console.log(`Pagination load: received ${data.history?.length || 0} messages at offset ${this.currentOffset}`);
            console.log('Has more messages:', data.hasMore);
            
            // Check if there are more messages
            this.hasMoreMessages = data.hasMore || false;
            
            if (data.history && data.history.length > 0) {
                console.log('Loading older messages...');
                data.history.forEach((msg, index) => {
                    console.log(`Older message ${index + 1}: "${msg.content}" at ${msg.timestamp}`);
                });
                
                // Store current scroll position before adding messages
                const previousScrollHeight = chatMessages.scrollHeight;

                // Insert older messages at the beginning while preserving chronological order.
                // data.history is expected to be ordered oldest -> newest for the chunk.
                // We iterate in reverse and insert each message before the first child so the
                // final order at the top remains oldest -> newest.
                for (let i = data.history.length - 1; i >= 0; i--) {
                    const msg = data.history[i];
                    const msgDiv = document.createElement('div');
                    msgDiv.classList.add('chat-message');

                    // Check if the message is from the current user or the other user
                    if (msg.from === currentUser?.nickname) {
                        msgDiv.classList.add('sent');       // message you sent
                    } else {
                        msgDiv.classList.add('received');   // message received
                    }

                    // Add the content
                    msgDiv.textContent = msg.content;

                    // Add timestamp
                    const timeDiv = document.createElement('div');
                    timeDiv.classList.add('chat-message-time');
                    timeDiv.textContent = new Date(msg.timestamp).toLocaleTimeString();
                    msgDiv.appendChild(timeDiv);

                    // Insert before the current first child so that order is preserved
                    chatMessages.insertBefore(msgDiv, chatMessages.firstChild);
                }
                
                // Hide loading indicator
                this.hideChatLoadingIndicator();
                
                // Maintain scroll position so user doesn't lose their place
                const newScrollHeight = chatMessages.scrollHeight;
                chatMessages.scrollTop = newScrollHeight - previousScrollHeight;
                
                // Update offset for next load
                this.currentOffset += data.history.length;
            } else if (this.currentOffset > 0) {
                // No more messages to load
                this.showNoMoreMessagesIndicator();
            }
            
            this.isLoadingHistory = false;
        })
        .catch(err => {
            console.error('Error fetching chat history:', err);
            // Hide loading indicator on error
            this.hideChatLoadingIndicator();
            this.isLoadingHistory = false;
        });
    }

    /**
     * Load more messages when scrolling to top
     */
    loadMoreMessages() {
        if (!this.currentChatUser || this.isLoadingHistory || !this.hasMoreMessages) {
            return;
        }
        
        console.log('Loading more messages for', this.currentChatUser);
        
        // Show loading indicator
        this.showChatLoadingIndicator();
        
        this.loadChatHistory(this.currentChatUser);
    }

    /**
     * Show loading indicator at the top of chat messages
     */
    showChatLoadingIndicator() {
        const chatMessages = document.getElementById('chat-messages');
        if (!chatMessages) return;
        
        // Remove existing loading indicator if present
        const existingLoader = chatMessages.querySelector('.chat-loading');
        if (existingLoader) {
            existingLoader.remove();
        }
        
        // Create loading indicator
        const loadingDiv = document.createElement('div');
        loadingDiv.className = 'chat-loading';
        loadingDiv.innerHTML = `
            <div class="spinner"></div>
            Loading more messages...
        `;
        
        // Insert at the beginning
        chatMessages.insertBefore(loadingDiv, chatMessages.firstChild);
    }

    /**
     * Hide loading indicator
     */
    hideChatLoadingIndicator() {
        const chatMessages = document.getElementById('chat-messages');
        if (!chatMessages) return;
        
        const loadingDiv = chatMessages.querySelector('.chat-loading');
        if (loadingDiv) {
            loadingDiv.remove();
        }
    }

    /**
     * Show "no more messages" indicator
     */
    showNoMoreMessagesIndicator() {
        const chatMessages = document.getElementById('chat-messages');
        if (!chatMessages) return;
        
        // Remove loading indicator first
        this.hideChatLoadingIndicator();
        
        // Check if indicator already exists
        const existing = chatMessages.querySelector('.no-more-messages');
        if (existing) return;
        
        // Create "no more messages" indicator
        const noMoreDiv = document.createElement('div');
        noMoreDiv.className = 'no-more-messages';
        noMoreDiv.style.cssText = `
            text-align: center;
            padding: 10px;
            color: #666;
            font-size: 0.8rem;
            opacity: 0.7;
            border-bottom: 1px solid #3a3a47;
            margin-bottom: 10px;
        `;
        noMoreDiv.textContent = '— Beginning of conversation —';
        
        // Insert at the beginning
        chatMessages.insertBefore(noMoreDiv, chatMessages.firstChild);
        
        // Auto-remove after 3 seconds
        setTimeout(() => {
            if (noMoreDiv.parentNode) {
                noMoreDiv.remove();
            }
        }, 3000);
    }

    /**
     * Get current chat user
     */
    getCurrentChatUser() {
        return this.currentChatUser;
    }

    /**
     * Clear chat state
     */
    clearState() {
        this.currentChatUser = null;
        this.closeChatWidget();
        this.disconnectWebSocket();
        
        // Reset pagination state
        this.currentOffset = 0;
        this.hasMoreMessages = true;
        this.isLoadingHistory = false;
    }

    /**
     * Handle user joined event
     */
    handleUserJoined(message) {
        console.log(`User ${message.content} joined`);
        
        // Update stored online users and refresh all users display
        if (message.online_users) {
            this.onlineUsers = message.online_users;
            this.app.loadAllUsers();
        }
        
        // Show notification if it's not the current user
        const currentUser = this.app.auth.getCurrentUser();
        if (message.content !== currentUser?.nickname) {
            this.app.ui.showToast(`🟢 ${message.content} joined the forum`, 'success', 4000);
        }
    }

    /**
     * Handle user left event
     */
    handleUserLeft(message) {
        console.log(`User ${message.content} left`);
        
        // Update stored online users and refresh all users display
        if (message.online_users) {
            this.onlineUsers = message.online_users;
            this.app.loadAllUsers();
        }
        
        // Show notification
        this.app.ui.showToast(`🔴 ${message.content} left the forum`, 'info', 4000);
        
        // Close chat if it was open with this user
        if (this.currentChatUser === message.content) {
            this.app.ui.showToast(`Chat with ${message.content} closed (user disconnected)`, 'warning', 3000);
            this.closeChatWidget();
        }
    }

    /**
     * Handle initial online users list
     */
    handleInitialOnlineUsers(message) {
        console.log('Received initial online users:', message.online_users);
        
        // Store online users and load all users for display
        if (message.online_users) {
            this.onlineUsers = message.online_users;
            console.log('Calling loadAllUsers from handleInitialOnlineUsers');
            this.app.loadAllUsers();
        }
    }

    /**
     * Handle online users update (triggered after new messages)
     */
    handleOnlineUsersUpdate(message) {
        console.log('Received online users update:', message.online_users);
        
        // Store online users for UI rendering
        if (message.online_users) {
            this.onlineUsers = message.online_users;
            this.app.loadAllUsers();
        }
    }

    /**
     * Get current online users from WebSocket
     */
    getOnlineUsers() {
        // This would contain the last received online users list
        return this.onlineUsers || [];
    }
}