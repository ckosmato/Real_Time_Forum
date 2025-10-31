 /**
 * Posts Module
 * Handles post creation, viewing, comments, categories, and post management
 */
class PostsManager {
    constructor(app) {
        this.app = app;
        this.categories = [];
        this.posts = [];
        this.selectedPostId = null;
    }

    /**
     * Initialize posts functionality
     */
    init() {
        this.bindPostEvents();
    }

    /**
     * Bind post-related event listeners
     */
    bindPostEvents() {
        // Post creation
        document.getElementById('createPostForm').addEventListener('submit', (e) => this.handleCreatePost(e));
    }

    /**
     * Set categories data
     */
    setCategories(categories) {
        this.categories = categories;
        this.renderCategories();
    }

    /**
     * Set posts data
     */
    setPosts(posts) {
        this.posts = posts;
        this.renderPosts();
    }

    /**
     * Get posts data
     */
    getPosts() {
        return this.posts;
    }

    /**
     * Render categories in the sidebar
     */
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
    }

    /**
     * Filter posts by category
     */
    async filterByCategory(categoryId) {
        // Update active category
        document.querySelectorAll('.category-item').forEach(item => {
            item.classList.remove('active');
        });
        event.target.classList.add('active');

        this.app.ui.showLoading();

        try {
            let url = '/dashboard';
            if (categoryId) {
                url = `/category/${categoryId}`;
            }

            // Get session id from cookie
            const sessionId = this.app.auth.getCookie('session_id');

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
                this.app.ui.showToast('Failed to load posts', 'error');
            }
        } catch (error) {
            this.app.ui.showToast('Network error', 'error');
        } finally {
            this.app.ui.hideLoading();
        }
    }

    /**
     * Render posts in the main content area
     */
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

    /**
     * Create a post card element
     */
    createPostCard(post) {
        const card = document.createElement('div');
        card.className = 'post-card';
        card.addEventListener('click', () => this.viewPost(post.ID)); // Changed from post.id to post.ID

        const categoriesHtml = post.Categories ? 
            post.Categories.map(cat => `<span class="category-tag">${cat}</span>`).join('') : '';

        card.innerHTML = `
            <div class="post-header">
                <h3 class="post-title">${this.app.ui.escapeHtml(post.Title)}</h3>
                <div class="post-meta">
                    <div>By: ${this.app.ui.escapeHtml(post.AuthorName)}</div>
                    <div>${this.app.ui.formatDate(post.CreatedAt)}</div>
                </div>
            </div>
            <div class="post-content">
                ${this.app.ui.escapeHtml((post.Content || '').substring(0, 200))}${(post.Content || '').length > 200 ? '...' : ''}
            </div>
            <div class="post-categories">
                ${categoriesHtml}
            </div>
        `;

        return card;
    }

    /**
     * View a specific post
     */
    async viewPost(postId) {
        console.log('ViewPost called with postId:', postId, 'type:', typeof postId);
        this.selectedPostId = postId;
        this.app.ui.showLoading();

        try {
            const url = `/post?id=${postId}`;
            console.log('Fetching URL:', url);
            const response = await fetch(url, { 
                credentials: 'same-origin',
                headers: {
                    'X-Session-ID': this.app.auth.getCookie('session_id'),
                    'Accept': 'application/json'
                }
            });
            
            console.log('ViewPost response status:', response.status);
            
            if (response.ok) {
                const data = await response.json();
                this.renderPostDetail(data.post, data.comments || []);
                this.app.ui.showView('post');
            } else {
                const data = await response.json();
                console.error('ViewPost error:', data);
                this.app.ui.showToast('Failed to load post: ' + (data.error || 'Unknown error'), 'error');
            }
        } catch (error) {
            console.error('ViewPost network error:', error);
            this.app.ui.showToast('Network error while loading post', 'error');
        } finally {
            this.app.ui.hideLoading();
        }
    }

    /**
     * Render post detail view
     */
    renderPostDetail(post, comments) {
        const container = document.getElementById('post-detail');
        
        const categoriesHtml = post.Categories ? 
            post.Categories.map(cat => `<span class="category-tag">${cat}</span>`).join('') : '';

        const commentsHtml = comments.map(comment => `
            <div class="comment">
                <div class="comment-header">
                    <span class="comment-author">${this.app.ui.escapeHtml(comment.AuthorName || comment.authorName)}</span>
                    <span class="comment-date">${this.app.ui.formatDate(comment.CreatedAt || comment.createdAt)}</span>
                </div>
                <div class="comment-content">${this.app.ui.escapeHtml(comment.Content || comment.content)}</div>
            </div>
        `).join('');

        const currentUser = this.app.auth.getCurrentUser();

        container.innerHTML = `
            <div class="post-navigation">
                <button onclick="app.ui.showView('home')" class="back-btn">← Back to Posts</button>
            </div>
            <div class="post-header">
                <h1 class="post-title">${this.app.ui.escapeHtml(post.Title)}</h1>
                <div class="post-meta">
                    <div>By: ${this.app.ui.escapeHtml(post.AuthorName)}</div>
                    <div>${this.app.ui.formatDate(post.CreatedAt)}</div>
                </div>
            </div>
            <div class="post-content" style="margin: 2rem 0;">
                ${this.app.ui.escapeHtml(post.Content || '').replace(/\n/g, '<br>')}
            </div>
            <div class="post-categories">
                ${categoriesHtml}
            </div>
            
            <div class="comments-section">
                <h3>Comments (${comments.length})</h3>
                
                ${currentUser ? `
                    <form class="comment-form" onsubmit="app.posts.handleCreateComment(event)">
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

    /**
     * Handle creating a new comment
     */
    async handleCreateComment(e) {
        e.preventDefault();
        
        const textarea = e.target.querySelector('textarea');
        const content = textarea.value.trim();
        
        if (!content) return;

        this.app.ui.showLoading();

        try {
            const formData = new FormData();
            formData.append('comment', content);
            formData.append('post_id', this.selectedPostId);

            const response = await fetch('/post/createcomment', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'X-Session-ID': this.app.auth.getCookie('session_id')
                },
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                this.app.ui.showToast('Comment posted successfully!', 'success');
                textarea.value = '';
                // Reload post to show new comment
                await this.viewPost(this.selectedPostId);
            } else {
                this.app.ui.showToast(data.error || 'Failed to post comment', 'error');
            }
        } catch (error) {
            this.app.ui.showToast('Network error', 'error');
        } finally {
            this.app.ui.hideLoading();
        }
    }

    /**
     * Handle creating a new post
     */
    async handleCreatePost(e) {
        e.preventDefault();
        this.app.ui.showLoading();

        const title = document.getElementById('post-title').value;
        const content = document.getElementById('post-content').value;
        const selectedCategories = Array.from(document.querySelectorAll('#post-categories input:checked'))
            .map(cb => cb.value);

        console.log('Creating post with:', { title, content, selectedCategories });

        if (selectedCategories.length === 0) {
            this.app.ui.showToast('Please select at least one category', 'error');
            this.app.ui.hideLoading();
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
                    'X-Session-ID': this.app.auth.getCookie('session_id')
                },
                body: formData
            });

            console.log('Response status:', response.status, 'Response ok:', response.ok);

            const data = await response.json();
            console.log('Response data:', data);

            if (response.ok) {
                this.app.ui.showToast('Post created successfully!', 'success');
                document.getElementById('createPostForm').reset();
                await this.app.loadDashboard(); // Refresh dashboard
                this.app.ui.showView('home');
            } else {
                this.app.ui.showToast(data.error || 'Failed to create post', 'error');
            }
        } catch (error) {
            console.error('Network error details:', error);
            this.app.ui.showToast('Network error', 'error');
        } finally {
            this.app.ui.hideLoading();
        }
    }

    /**
     * Load user's own posts
     */
    async loadMyPosts() {
        const currentUser = this.app.auth.getCurrentUser();
        if (!currentUser) return;
        
        this.app.ui.showLoading();

        try {
            // This would need a new endpoint in your backend
            const response = await fetch(`/dashboard/my-posts`, {
                method: 'GET',
                headers: {
                    'X-SESSION-ID': this.app.auth.getCookie('session_id')
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
                this.app.ui.showToast('Failed to load your posts', 'error');
            }
        } catch (error) {
            this.app.ui.showToast('Network error', 'error');
        } finally {
            this.app.ui.hideLoading();
        }
    }

    /**
     * Get selected post ID
     */
    getSelectedPostId() {
        return this.selectedPostId;
    }

    /**
     * Clear posts state
     */
    clearState() {
        this.posts = [];
        this.selectedPostId = null;
        
        // Clear displayed content
        const postsContainer = document.getElementById('posts-container');
        const myPostsContainer = document.getElementById('my-posts-container');
        
        if (postsContainer) postsContainer.innerHTML = '';
        if (myPostsContainer) myPostsContainer.innerHTML = '';
    }
}