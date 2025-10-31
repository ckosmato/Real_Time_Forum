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
                this.displayChatMessage(message);
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
            
            // Set the chat user in UI
            chatUsername.innerHTML = `<i class="fa-solid fa-comment"></i> Chat with ${this.app.ui.escapeHtml(username)}`;
            
            // Show the chat widget
            chatWidget.style.display = 'flex';
            
            // Focus on the input
            const chatInput = document.getElementById('chat-input');
            if (chatInput) {
                chatInput.focus();
            }
            
            this.loadChatHistory(username);
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
        }
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
     * Load chat history with a user
     */
    loadChatHistory(username) {
        console.log("Calling loadChatHistory for", username);
        const chatMessages = document.getElementById('chat-messages');

        fetch(`/chathistory?user2=${encodeURIComponent(username)}`, {
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
            chatMessages.innerHTML = '';

            const currentUser = this.app.auth.getCurrentUser();
            
            data.history.forEach(msg => {
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

                // Optionally, add a timestamp
                const timeDiv = document.createElement('div');
                timeDiv.classList.add('chat-message-time');
                timeDiv.textContent = new Date(msg.timestamp).toLocaleTimeString();
                msgDiv.appendChild(timeDiv);

                chatMessages.appendChild(msgDiv);
            });

            // Scroll to bottom
            chatMessages.scrollTop = chatMessages.scrollHeight;
        })
        .catch(err => {
            console.error('Error fetching chat history:', err);
            chatMessages.innerHTML = '<div class="chat-message error">Failed to load chat history.</div>';
        });
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
    }
}