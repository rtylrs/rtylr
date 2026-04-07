// Bot detection and security
(function () {
    const botPatterns = [
        /bot/i, /crawler/i, /spider/i, /scraper/i, /scraping/i,
        /GPTBot/i, /CCBot/i, /anthropic-ai/i, /Claude-Web/i, /Omgilibot/i,
        /ai/i, /crawler/i, /spider/i, /scraper/i, /scraping/i,
        /googlebot/i, /bingbot/i, /slurp/i, /ia_archiver/i,
        /archive\.org_bot/i, /twitterbot/i, /facebookexternalhit/i, /linkedinbot/i
    ];

    const userAgent = navigator.userAgent;
    const isBot = botPatterns.some(pattern => pattern.test(userAgent));

    if (isBot) {
        document.body.innerHTML = '<h1>Access Denied</h1><p>Bot access is not allowed.</p>';
        throw new Error('Bot access blocked');
    }
})();

// Global variables  
const BASE_URL = 'http://localhost:9999';
let currentSession = null;
let currentUserId = 1; // Default user ID
let isStreaming = false;

// Initialize session on page load
window.addEventListener('load', initializeSession);

async function initializeSession() {
    try {
        console.log('Attempting to create session...');
        const response = await fetch(`${BASE_URL}/api/memory/sessions`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                user_id: currentUserId,
                title: `Chat Session - ${new Date().toLocaleString()}`
            })
        });

        console.log('Response status:', response.status);
        console.log('Response headers:', response.headers);

        const data = await response.json();
        console.log('Response data:', data);

        if (data.success) {
            currentSession = data.session;
            updateSessionInfo();
            addSystemMessage(`Session created: ${currentSession.session_id}`);
        } else {
            const errorMsg = typeof data.error === 'object' ? JSON.stringify(data.error) : (data.error || 'Unknown error');
            addSystemMessage(`Failed to create session: ${errorMsg}. Memory features disabled.`);
        }
    } catch (error) {
        console.error('Failed to initialize session:', error);
        addSystemMessage(`Failed to initialize memory: ${error.message}. Continuing without memory features.`);
    }
}

function updateSessionInfo() {
    const sessionInfo = document.getElementById('sessionInfo');
    if (currentSession) {
        sessionInfo.textContent = `Session: ${currentSession.session_id.slice(-8)}`;
    } else {
        sessionInfo.textContent = 'Session: Error';
    }
}

function autoResize(textarea) {
    // Disabled - keep fixed height
    // textarea.style.height = 'auto';
    // textarea.style.height = Math.min(textarea.scrollHeight, 120) + 'px';
}

function addMessage(content, isUser = false, hasMemory = false) {
    const chatMessages = document.getElementById('chatMessages');
    const messageDiv = document.createElement('div');
    messageDiv.className = `message ${isUser ? 'user' : 'assistant'}`;

    if (isUser) {
        // User message with gray bubble, no header
        messageDiv.innerHTML = `<div class="message-content">${content}</div>`;
    } else {
        // Assistant message with markdown formatting
        const formattedContent = parseMarkdown(content);
        messageDiv.innerHTML = `<div class="message-content">${formattedContent}</div>`;
    }

    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

function addSystemMessage(content) {
    const chatMessages = document.getElementById('chatMessages');
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message system';

    messageDiv.innerHTML = `
        <div class="message-content">${content}</div>
    `;

    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

function addTypingIndicator() {
    const chatMessages = document.getElementById('chatMessages');
    const typingDiv = document.createElement('div');
    typingDiv.className = 'message assistant';
    typingDiv.id = 'typingIndicator';
    typingDiv.innerHTML = `
        <div class="typing-indicator">
            <div class="typing-dot"></div>
            <div class="typing-dot"></div>
            <div class="typing-dot"></div>
        </div>
    `;

    chatMessages.appendChild(typingDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

function removeTypingIndicator() {
    const typingIndicator = document.getElementById('typingIndicator');
    if (typingIndicator) {
        typingIndicator.remove();
    }
}

async function getContext() {
    if (!currentSession) return '';

    try {
        // Send POST request with session ID and user ID in body to match the ContextRequest struct
        const response = await fetch(`${BASE_URL}/api/memory/context`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                session_id: currentSession.session_id,
                user_id: currentUserId
            })
        });

        if (!response.ok) {
            const errorText = await response.text();
            console.error('Get context failed:', response.status, errorText);
            return '';
        }

        const data = await response.json();
        console.log('Context response:', data);

        if (data.success && data.context) {
            let contextString = '\n\n[Previous conversation context]:\n';
            console.log('Processing context:', data.context);

            // Process messages
            if (data.context.messages && data.context.messages.length > 0) {
                console.log('Found messages:', data.context.messages);
                data.context.messages.forEach(message => {
                    console.log('Message:', message);
                    contextString += `${message.role}: ${message.content}\n`;
                });
            }

            // Process memories  
            if (data.context.memories && data.context.memories.length > 0) {
                console.log('Found memories:', data.context.memories);
                data.context.memories.forEach(memory => {
                    contextString += `[Memory] ${memory.memory_type}: ${memory.value}\n`;
                });
            }

            console.log('Final context string:', contextString);
            return contextString;
        }
    } catch (error) {
        console.error('Failed to get context:', error);
    }
    return '';
}

// Markdown parser for formatting AI responses
function parseMarkdown(text) {
    // Convert markdown to HTML
    return text
        // Bold text: **text** -> <strong>text</strong>
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        // Italic text: *text* -> <em>text</em>
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        // Inline code: `code` -> <code>code</code>
        .replace(/`([^`]+)`/g, '<code>$1</code>')
        // Line breaks
        .replace(/\n/g, '<br>')
        // Lists: * item -> <li>item</li>
        .replace(/^\* (.+)$/gm, '<li>$1</li>')
        // Wrap lists in <ul> tags
        .replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>');
}

// Function to safely append HTML content
function appendFormattedContent(element, text) {
    const formattedText = parseMarkdown(text);
    element.innerHTML += formattedText;
}

async function sendMessage() {
    const promptInput = document.getElementById('promptInput');
    const sendButton = document.getElementById('sendButton');
    const prompt = promptInput.value.trim();

    if (!prompt || isStreaming) return;

    // Add user message to chat
    addMessage(prompt, true);


    // Clear input and disable button
    promptInput.value = '';
    promptInput.style.height = 'auto';
    sendButton.disabled = true;
    isStreaming = true;

    // Show typing indicator while preparing request
    addTypingIndicator();

    try {
        // Get conversation context
        const context = await getContext();
        const enhancedPrompt = prompt + context;

        console.log('Sending AI request:', {
            prompt: enhancedPrompt,
            session_id: currentSession ? currentSession.session_id : "",
            user_id: currentUserId
        });

        // Remove typing indicator and create AI response message
        removeTypingIndicator();

        const messageDiv = document.createElement('div');
        messageDiv.className = 'message assistant';
        const uniqueId = 'streamingContent_' + Date.now();
        messageDiv.innerHTML = `<div class="message-content" id="${uniqueId}"></div>`;

        document.getElementById('chatMessages').appendChild(messageDiv);
        document.getElementById('chatMessages').scrollTop = document.getElementById('chatMessages').scrollHeight;

        const streamingContent = document.getElementById(uniqueId);
        let fullResponse = '';

        // Stream response from AI using Server-Sent Events
        const response = await fetch(`${BASE_URL}/api/ai/generate-text-stream`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                prompt: enhancedPrompt,
                session_id: currentSession ? currentSession.session_id : "",
                user_id: currentUserId
            })
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        // Handle SSE response manually since we need to send POST data
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        async function readStream() {
            return reader.read().then(async ({ done, value }) => {
                if (done) {
                    return;
                }

                buffer += decoder.decode(value, { stream: true });
                const lines = buffer.split('\n');

                // Keep the last incomplete line in the buffer
                buffer = lines.pop() || '';

                for (const line of lines) {
                    if (line.trim() === '') continue;

                    if (line.startsWith('data: ')) {
                        const data = line.substring(6);
                        if (data.trim() === '') continue;

                        try {
                            const token = JSON.parse(data);

                            if (token.error) {
                                streamingContent.textContent += `\n[ERROR] ${token.error}`;
                                continue;
                            }

                            if (token.done) {
                                continue;
                            }

                            if (token.text) {
                                console.log('Received token:', token.text);
                                fullResponse += token.text;

                                // Stream the text character by character for smooth appearance
                                for (let i = 0; i < token.text.length; i++) {
                                    // Parse markdown for the accumulated text so far
                                    const currentText = fullResponse.substring(0, fullResponse.length - token.text.length + i + 1);
                                    const formattedText = parseMarkdown(currentText);
                                    streamingContent.innerHTML = formattedText;

                                    document.getElementById('chatMessages').scrollTop = document.getElementById('chatMessages').scrollHeight;

                                    // Small delay between characters for smooth streaming
                                    await new Promise(resolve => setTimeout(resolve, 20));
                                }
                            }
                        } catch (error) {
                            console.warn('Failed to parse SSE data:', data, error);
                        }
                    }
                }

                return readStream();
            });
        }

        await readStream();

    } catch (error) {
        removeTypingIndicator();
        if (error.name === 'TypeError' && error.message.includes('Failed to fetch')) {
            addMessage('❌ Connection failed. Please check if the server is running.');
        } else {
            addMessage(`❌ Error: ${error.message}`);
        }
    } finally {
        sendButton.disabled = false;
        isStreaming = false;
        promptInput.focus();
    }
}

// Event listeners
document.getElementById('promptInput').addEventListener('keydown', function (e) {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        sendMessage();
    }
});

// Focus input on page load
window.addEventListener('load', () => {
    document.getElementById('promptInput').focus();
});
