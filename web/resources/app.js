import {css, html, LitElement} from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';

import './molecules.js';

// ChatApp Component
class ChatApp extends LitElement {

    static properties = {
        messages: {type: Array},
        processing: {},
        eventSource: {type: Object},
        contextId: {type: String}
    };

    static styles = css`
        :host {
            display: flex;
            flex-direction: column;
            height: 100vh;
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 0;
        }

        .chat-container {
            flex: 1;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        .chat-container.mode-init {
            justify-content: space-around;
        }
        .chat-container.mode-chat {
            justify-content: flex-start;
        }

        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 10px;
            background-color: #fff;
            border-bottom: 1px solid #ddd;
            gap: 10px;
            position: sticky;
            top: 0;
            z-index: 100;
        }

        .context-title {
            font-weight: bold;
            font-size: 16px;
        }

        .header-actions {
            display: flex;
            gap: 10px;
        }

        .messages {
            flex: 1;
            overflow-y: auto;
            padding: 10px;
            background-color: white;
            border-bottom: 1px solid #ddd;
        }

        .empty-messages {
            flex: 1;
            display: flex;
            justify-content: center;
            align-items: center;
            background-color: white;
        }
        
        .input-area {
            padding: 2rem;
        }
    `;

    constructor() {
        super();
        this.messages = [];
        this.processing = false;
        this.needScoll = false;
        this.eventSource = null;
        this.contextId = 'default';
    }

    connectedCallback() {
        super.connectedCallback();
        this.setupStream();
    }

    disconnectedCallback() {
        super.disconnectedCallback();
        if (this.eventSource) {
            this.eventSource.close();
        }
    }


    updated(changedProperties) {
        if (this.needScoll) {
            this.needScoll = false
            this.scrollToBottom()
        }
        // Wenn sich der ausgewählte Kontext ändert, den Stream neu einrichten
        if (changedProperties.has('contextId')) {
            this.setupStream();
        }
    }

    async cancelChat() {
        try {
            const response = await fetch(`http://localhost:8080/chat/${this.contextId}/cancel`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            });
            if (!response.ok) {
                throw new Error('Failed to cancel chat');
            }
            await this.loadChatContext();
        } catch (error) {
            console.error('Error canceling chat:', error);
        }
    }

    async compactChat() {
        this.isLoading = true;
        try {
            const response = await fetch(`http://localhost:8080/chat/${this.contextId}/compact`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            });
            if (!response.ok) {
                throw new Error('Failed to compact chat');
            }
            await this.loadChatContext();
        } catch (error) {
            console.error('Error compacting chat:', error);
        } finally {
            this.isLoading = false;
        }
    }

    async clearChat() {
        try {
            const response = await fetch(`http://localhost:8080/chat/${this.contextId}/clear`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            });
            if (!response.ok) {
                throw new Error('Failed to clear chat');
            }
            this.messages = []
        } catch (error) {
            console.error('Error clearing chat:', error);
        } finally {
            this.isLoading = false;
        }
    }


    setupStream() {
        // Schließen des bestehenden EventSource, falls vorhanden
        if (this.eventSource) {
            this.eventSource.close();
        }

        this.eventSource = new EventSource(`http://localhost:8080/chat/${this.contextId}/context`);
        this.eventSource.onfinish = () => console.warn("finish")

        this.eventSource.onmessage = (event) => {
            console.warn("UNKNOWN EVENT", event);
        }

        this.eventSource.addEventListener("state-change", (event) => {
            if (event.data === "processing") {
                this.processing = true
            } else if (event.data === "idle") {
                this.processing = false
            }
        });

        this.eventSource.addEventListener("chat-message", (event) => {
            console.info(event.data)
            const message = JSON.parse(event.data);

            let lastMessage = this.messages.length - 1;
            const lastRole = this.messages.length > 0 ? this.messages[lastMessage].role : undefined

            if (lastRole !== message.role) {
                console.warn("NEW MESSAGE!!!", message)
                lastMessage++
                this.messages = [...this.messages, message];
            } else {
                console.warn("APPEND MESSAGE!!! lastMessage", message)
                this.messages[lastMessage] = {
                    ...this.messages[lastMessage],
                    ...message,
                    content: (this.messages[lastMessage].content || '') + (message.content || ''),
                    tool_calls: [].concat(this.messages[lastMessage].tool_calls || []).concat(message.tool_calls || []),
                }
            }
            setTimeout(() => {
                this.requestUpdate();
                this.scrollToBottom();
            })
        });

        this.eventSource.onerror = (error) => {
            console.error('EventSource failed:', error);
            this.eventSource.close();
        };
    }

    allMessagesAreSystem() {
        return this.messages.every(m => m.role === 'system')
    }



    render() {
        const initial = this.allMessagesAreSystem()
        return html`
            <div class="chat-container mode-${initial ? 'init' : 'chat'}">
                ${initial ? undefined : html`
                    <div class="header">
                        <div class="context-title">Context: ${this.contextId}</div>
                        <div class="header-actions">
                            <atomic-button label="Cancel" ?disabled=${!this.processing} @button-click=${this.cancelChat}></atomic-button>
                            <atomic-button label="Compact" ?disabled=${this.processing} @button-click=${this.compactChat}></atomic-button>
                            <atomic-button label="Clear" ?disabled=${this.processing} @button-click=${this.clearChat}></atomic-button>
                        </div>
                    </div>
                    <div class="messages">
                        ${this.messages.map(msg => html`
                            <chat-message .message=${msg}></chat-message>`)}
                    </div>
                `}
                <div class="input-area">
                    ${this.processing ? html`<pre>processing</pre>` : null}
                    ${initial ? html`<h3>Was geht up?</h3>` : undefined}
                    <chat-input .processing=${this.processing}
                                @send-message=${this.sendMessage}
                    ></chat-input>
                </div>
            </div>
        `;
    }

    scrollToBottom() {
        const messagesContainer = this.shadowRoot.querySelector('.messages');
        if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
    }

    async sendMessage(e) {
        try {
            const response = await fetch(`http://localhost:8080/chat/${this.contextId}/message`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({message: e.detail.message}),
            });
            if (!response.ok) {
                throw new Error('Failed to send message');
            }
        } catch (error) {
            console.error('Error:', error);
        }
    }
}

customElements.define('chat-app', ChatApp);