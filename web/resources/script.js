import { LitElement, html, css } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';

class ChatApp extends LitElement {
    static properties = {
        messages: { type: Array },
        inputValue: { type: String },
        isLoading: { type: Boolean },
    };

    constructor() {
        super();
        this.messages = [];
        this.inputValue = '';
        this.isLoading = false;
    }

    static styles = css`
        :host {
            display: flex;
            flex-direction: column;
            height: 100vh;
            font-family: Arial, sans-serif;
            background-color: #f5f5f5;
            margin: 0;
            padding: 0;
        }

        .chat-container {
            flex: 1;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }

        .messages {
            flex: 1;
            overflow-y: auto;
            padding: 10px;
            background-color: white;
            border-bottom: 1px solid #ddd;
        }

        .message {
            margin-bottom: 10px;
            padding: 10px;
            border-radius: 5px;
            background-color: #e3f2fd;
        }

        .input-area {
            display: flex;
            padding: 10px;
            background-color: #fff;
            border-top: 1px solid #ddd;
        }

        input {
            flex: 1;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 16px;
        }

        button {
            margin-left: 10px;
            padding: 10px 20px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
        }

        button:hover {
            background-color: #0056b3;
        }
    `;

    render() {
        return html`
            <div class="chat-container">
                <div class="messages">
                    ${this.messages.map(msg => html`<div class="message">${msg}</div>`)}
                </div>
                <div class="input-area">
                    <input
                        .value=${this.inputValue}
                        @input=${this.handleInput}
                        @keyup=${this.handleKeyUp}
                        placeholder="Type your message here..."
                    >
                    <button @click=${this.sendMessage}>Send</button>
                </div>
            </div>
        `;
    }

    handleInput(e) {
        this.inputValue = e.target.value;
    }

    handleKeyUp(e) {
        if (e.key === 'Enter') {
            this.sendMessage();
        }
    }

    async sendMessage() {
        if (this.inputValue.trim() === '') return;

        const message = this.inputValue;
        this.inputValue = '';
        this.isLoading = true;

        // Add user message to the chat
        this.messages = [...this.messages, `You: ${message}`];
        this.requestUpdate();

        try {
            // Send message to the REST API
            const response = await fetch('http://localhost:8080/api/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ message }),
            });

            if (!response.ok) {
                throw new Error('Failed to send message');
            }

            // Handle SSE stream
            const eventSource = new EventSource('http://localhost:8080/api/chat/stream');
            eventSource.onmessage = (event) => {
                const data = JSON.parse(event.data);
                this.messages = [...this.messages, `Bot: ${data.message}`];
                this.requestUpdate();
            };

            eventSource.onerror = () => {
                eventSource.close();
                this.isLoading = false;
            };
        } catch (error) {
            console.error('Error:', error);
            this.messages = [...this.messages, `Error: ${error.message}`];
            this.requestUpdate();
            this.isLoading = false;
        }
    }
}

customElements.define('chat-app', ChatApp);