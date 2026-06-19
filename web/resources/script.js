
import { LitElement, html, css } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';

import { marked } from "https://cdn.jsdelivr.net/npm/marked/lib/marked.esm.js";
import DOMPurify from "https://cdn.jsdelivr.net/npm/dompurify@3.0.8/dist/purify.es.mjs";


class MarkdownView extends LitElement {
    static properties = {
        content: { type: String },
    };

    constructor() {
        super();
        this.content = "# Hallo\n**Markdown sicher gerendert**";
    }

    static styles = css`
    .md {
      line-height: 1.5;
    }

    .md pre {
      padding: 10px;
      overflow: auto;
    }
  `;

    render() {
        const rawHtml = marked.parse(this.content || "");
        const safeHtml = DOMPurify.sanitize(rawHtml);
        return html`
          <div class="md" .innerHTML=${safeHtml}></div>
        `;
    }
}

customElements.define("markdown-view", MarkdownView);
// Message Component
class Message extends LitElement {
    static properties = {
        message: { type: Object },
    };

    constructor() {
        super();
        this.collapsed = false;
        this.message = {}
    }

    connectedCallback() {
        super.connectedCallback();
        // Standardmäßig ausgeklappt für 'assistant' und 'user', sonst eingeklappt
        this.collapsed = !(this.message.role === 'assistant' || this.message.role === 'user');
    }

    static styles = css`
        .message {
            margin-bottom: 12px;
            border-radius: 8px;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
            overflow: hidden;
            transition: all 0.2s ease;
        }

        .message-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 10px 15px;
            cursor: pointer;
            font-weight: 600;
            background-color: #f8f9fa;
            border-bottom: 1px solid #e9ecef;
        }

        .message-header:hover {
            background-color: #e9ecef;
        }

        .message-role {
            color: #495057;
        }

        .message-toggle {
            color: #6c757d;
            font-size: 14px;
        }

        .message-content {
            padding: 15px;
            display: none;
        }

        .message-content.visible {
            display: block;
        }

        .message-assistant {
            background-color: #f8f9ff;
            border-left: 4px solid #6b72ff;
        }

        .message-user {
            background-color: #fff8f8;
            border-left: 4px solid #ff6b6b;
        }

        .message-other {
            background-color: #f8f9fa;
            border-left: 4px solid #ced4da;
        }

        .tool-calls {
            margin-top: 10px;
            padding: 10px;
            background-color: #f8f9fa;
            border-radius: 5px;
            border-left: 3px solid #adb5bd;
        }

        .tool-calls-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            cursor: pointer;
            font-weight: 600;
            color: #495057;
        }

        .tool-call {
            margin-top: 8px;
            padding: 8px;
            background-color: #fff;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            color: #495057;
            border-left: 2px solid #ced4da;
        }

        .tool-call-details {
            display: none;
            margin-top: 5px;
            padding-left: 10px;
        }

        .tool-call.visible .tool-call-details {
            display: block;
        }
    `;

    toggleCollapse() {
        this.collapsed = !this.collapsed;
        this.requestUpdate();
    }

    renderToolCalls() {
        if (!this.message.tool_calls || this.message.tool_calls.length === 0) {
            return html``;
        }

        return html`
            <div class="tool-calls">
                <h4>Tool Calls:</h4>
                ${this.message.tool_calls.map(toolCall => html`
                    <div class="tool-call">
                        <strong>ID:</strong> ${toolCall.id}<br>
                        <strong>Type:</strong> ${toolCall.type}<br>
                        <strong>Function:</strong> ${toolCall.function.name}<br>
                        <strong>Arguments:</strong> ${toolCall.function.arguments}
                    </div>
                `)}
            </div>
        `;
    }

    render() {
        return html`
            <div class="message">
                <div class="message-header" @click=${this.toggleCollapse}>
                    <span>${this.message.role}</span>
                    <span>${this.message.collapsed ? '▶' : '▼'}</span>
                </div>
                <div class="message-content ${this.collapsed ? '' : 'visible'}">
                    <markdown-view .content=${this.message.content}></markdown-view>
                    ${this.renderToolCalls()}
                </div>
            </div>
        `;
    }
}

customElements.define('chat-message', Message);

// InputBar Component
class InputBar extends LitElement {
    static properties = {
        value: { type: String },
    };

    static styles = css`
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

    constructor() {
        super();
        this.value = '';
    }

    render() {
        return html`
            <div class="input-area">
                <input
                    .value=${this.value}
                    @input=${this.handleInput}
                    @keyup=${this.handleKeyUp}
                    placeholder="Type your message here..."
                >
                <button @click=${this.handleSend}>Send</button>
            </div>
        `;
    }

    handleInput(e) {
        this.value = e.target.value;
        this.dispatchEvent(new CustomEvent('input-change', { detail: { value: this.value } }));
    }

    handleKeyUp(e) {
        if (e.key === 'Enter') {
            this.handleSend();
        }
    }

    handleSend() {
        if (this.value.trim() !== '') {
            this.dispatchEvent(new CustomEvent('send-message', { detail: { message: this.value } }));
            this.value = '';
        }
    }
}

customElements.define('chat-input', InputBar);

// ChatApp Component
class ChatApp extends LitElement {

    connectedCallback() {
        super.connectedCallback();
        this.loadChatContext();
    }

    static properties = {
        messages: { type: Array },
        isLoading: { type: Boolean },
    };

    constructor() {
        super();
        this.messages = [];
        this.isLoading = false;
        this.needScoll = false;
    }

    updated(changedProperties) {
        if (this.needScoll) {
            this.needScoll = false
            this.scrollToBottom()
        }
    }

    async loadChatContext() {
        this.needScoll = true
        try {
            const response = await fetch('http://localhost:8080/chat/default/context', {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            });
            if (!response.ok) {
                throw new Error('Failed to load chat context');
            }
            const context = await response.json();
            this.messages = context.messages || [];
            setTimeout(() => this.scrollToBottom())
        } catch (error) {
            console.error('Error loading chat context:', error);
        }
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
    `;

    render() {
        return html`
            <div class="chat-container">
                <div class="messages">
                    ${this.messages.map(msg => html`<chat-message .message=${msg}></chat-message>`)}
                </div>
                <chat-input @send-message=${this.sendMessage} @input-change=${this.handleInputChange}></chat-input>
            </div>
        `;
    }

    handleInputChange(e) {
        this.inputValue = e.detail.value;
    }

    scrollToBottom() {
        const messagesContainer = this.shadowRoot.querySelector('.messages');
        if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
    }

    async sendMessage(e) {
        const message = e.detail.message;
        this.isLoading = true;

        // Add user message to the chat
        this.messages = [...this.messages, { role: 'user', content: message }];
        this.requestUpdate();
        this.scrollToBottom();

        try {
            // Send message to the REST API
            const response = await fetch('http://localhost:8080/chat/default/message', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ message: message }),
            });

            if (!response.ok) {
                throw new Error('Failed to send message');
            }

            // Add a placeholder message for the assistant's response
            const assistantMessageIndex = this.messages.length;
            this.messages = [...this.messages, { role: 'assistant', content: '' }];
            this.requestUpdate();
            this.scrollToBottom();

            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            var  buffer = "";
            while (true) {
                const {done, value} = await reader.read();
                if (done) break;

                buffer += decoder.decode(value, {stream: true});

                let parts = buffer.split("\n\n");
                buffer = parts.pop(); // letzter evtl. unvollständiger Block bleibt drin

                for (const part of parts) {
                    const line = part
                        .split("\n")
                        .find(l => l.startsWith("data:"));

                    if (!line) continue;

                    const data = line.replace("data: ", "").trim();
                    if (!data) continue;

                    if (data === "[DONE]") continue;

                    const json = JSON.parse(data);

                    this.messages[assistantMessageIndex] = {
                        ...this.messages[assistantMessageIndex],
                        content: this.messages[assistantMessageIndex].content += json.content
                    };
                }
                this.requestUpdate();
                this.scrollToBottom();
            }
        } catch (error) {
            console.error('Error:', error);
            this.messages = [...this.messages, { sender: 'Error', text: error.message }];
            this.requestUpdate();
            this.scrollToBottom();
        } finally {
            this.isLoading = false;
        }
    }
}

customElements.define('chat-app', ChatApp);