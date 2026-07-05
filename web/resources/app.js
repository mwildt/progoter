import {css, html, LitElement} from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';

// ChatApp Component
class ChatApp extends LitElement {

    connectedCallback() {
        super.connectedCallback();
        // this.loadChatContext();
        this.setupStream();
    }

    async cancelChat() {
        try {
            const response = await fetch('http://localhost:8080/chat/default/cancel', {
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
            const response = await fetch('http://localhost:8080/chat/default/compact', {
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
            const response = await fetch('http://localhost:8080/chat/default/clear', {
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

    static properties = {
        messages: {type: Array},
        processing: {},
        eventSource: {type: Object},
    };

    constructor() {
        super();
        this.messages = [];
        this.processing = false;
        this.needScoll = false;
        this.eventSource = null;
    }

    updated(changedProperties) {
        if (this.needScoll) {
            this.needScoll = false
            this.scrollToBottom()
        }
    }

    setupStream() {
        this.eventSource = new EventSource('http://localhost:8080/chat/default/context');
        this.eventSource.onfinish =  () => console.warn("finish")

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
                this.messages = [...this.messages, message] ;
            } else {
                console.warn("APPEND MESSAGE!!! lastMessage",  message)
                this.messages[lastMessage] = {
                    ...this.messages[lastMessage],
                    ... message,
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

    disconnectedCallback() {
        super.disconnectedCallback();
        if (this.eventSource) {
            this.eventSource.close();
        }
    }

    allMessagesAreSystem() {
        return this.messages.every(m => m.role === 'system')
    }

    static styles = css`
        :host {
            max-width: 1000px;
            min-width: 1000px;
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
            justify-content: flex-end;
            padding: 10px;
            background-color: #fff;
            border-bottom: 1px solid #ddd;
            gap: 10px;
            position: sticky;
            top: 0;
            z-index: 100;
        }

        .header button {
            padding: 8px 16px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
        }

        .header button:hover {
            background-color: #0056b3;
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

    render() {
        const initial = this.allMessagesAreSystem()
        return html`
            <div class="chat-container mode-${initial ? 'init' : 'chat'}">
                ${initial ? undefined : html`
                    <div class="header">
                        <button ?disabled=${!this.processing} @click=${this.cancelChat}>Cancel</button>
                        <button ?disabled=${this.processing} @click=${this.compactChat}>Compact</button>
                        <button ?disabled=${this.processing} @click=${this.clearChat}>Clear</button>
                    </div>
                    <div class="messages">
                        ${this.messages.map(msg => html`
                            <chat-message .message=${msg}></chat-message>`)}
                    </div>
                `}
                <div class="input-area">
                    ${this.processing ? html`<pre>processing</pre>`:null}
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
            const response = await fetch('http://localhost:8080/chat/default/message', {
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
