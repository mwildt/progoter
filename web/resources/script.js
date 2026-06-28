import {css, html, LitElement} from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';

import {marked} from "https://cdn.jsdelivr.net/npm/marked/lib/marked.esm.js";
import DOMPurify from "https://cdn.jsdelivr.net/npm/dompurify@3.0.8/dist/purify.es.mjs";

class AppLayout extends LitElement {

    constructor() {
        super();
    }

    static styles = css`
    :host {
        display: grid;
        justify-content: center;
    }
  `;

    render() {
        return html`
            <slot></slot>
        `;
    }
}

customElements.define("app-layout", AppLayout);

class MarkdownView extends LitElement {
    static properties = {
        content: {type: String},
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
        message: {type: Object},
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

        .message-teaser {
            color: #6c757d;
            font-size: 12px;
            margin-left: 8px;
            flex: 1;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .tool-call-label {
            color: #007bff;
            font-size: 12px;
            margin-left: 8px;
            font-weight: 600;
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

        .use-info {
            margin-top: 10px;
            padding: 10px;
            background-color: #f8f9fa;
            border-radius: 5px;
            border-left: 3px solid #adb5bd;
        }

        .use-info pre {
            margin: 0;
            white-space: pre-wrap;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            color: #495057;
        }

        .usage-info {
            margin-left: 10px;
            font-size: 12px;
            color: #6c757d;
            font-weight: normal;
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

    renderUseInfo() {
        if (!this.message.use_info) {
            return html``;
        }

        return html`
            <div class="use-info">
                <h4>Use Info:</h4>
                <pre>${JSON.stringify(this.message.use_info, null, 2)}</pre>
            </div>
        `;
    }

    renderUsageInfo() {
        if (!this.message.usage) {
            return html``;
        }

        return html`
            <span class="usage-info">
                Tokens: ${this.message.usage.total_tokens} (Prompt: ${this.message.usage.prompt_tokens}, Completion: ${this.message.usage.completion_tokens})
            </span>
        `;
    }

    render() {
        const teaser = this.collapsed && this.message.content ? this.message.content.substring(0, 50) + '...' : '';
        let labels = [];
        if (this.message.tool_calls && this.message.tool_calls.length > 0) {
            const toolCall = this.message.tool_calls[0];
            labels.push(`${toolCall.function.name}#${toolCall.id}`);
        }
        if (this.message.tool_call_id) {
            labels.push(`#${this.message.tool_call_id}`);
        }
        return html`
            <div class="message">
                <div class="message-header" @click=${this.toggleCollapse}>
                    <span>${this.message.role}</span>
                    ${labels.map(label => html`<span class="tool-call-label">${label}</span>`)}
                    <span class="message-teaser">${teaser}</span>
                    ${this.renderUsageInfo()}
                    <span>${this.collapsed ? '▶' : '▼'}</span>
                </div>
                <div class="message-content ${this.collapsed ? '' : 'visible'}">
                    <markdown-view .content=${this.message.content}></markdown-view>
                    ${this.renderToolCalls()}
                    ${this.renderUseInfo()}
                </div>
            </div>
        `;
    }
}

customElements.define('chat-message', Message);

// InputBar Component
class InputBar extends LitElement {
    static properties = {
        value: {type: String},
        processing: {}
    };

    static styles = css`
        .input-area {
            display: flex;
            background-color: #fff;
        }

        textarea {
              flex: 1 1 0%;
              padding: 8px;
              border: 1px solid rgb(221, 221, 221);
              border-radius: 4px;
              resize: none;
              line-height: 1.2em;
              min-height: 1.2em;
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
        this.processing = false;
    }

    render() {
        return html`
            <div class="input-area">
                <textarea
                        .value=${this.value}
                        @input=${this.handleInput}
                        @keyup=${this.handleKeyUp}
                        placeholder="Type your message here..." ></textarea>
                <button ?disabled=${this.processing} @click=${this.handleSend}>Send</button>
            </div>
        `;
    }

    handleInput(e) {
        this.value = e.target.value;
        this.dispatchEvent(new CustomEvent('input-change', {detail: {value: this.value}}));
        this.adjustTextareaHeight(e.target);
    }

    adjustTextareaHeight(textarea) {
        // Setze die Höhe auf 'auto', um die tatsächliche Höhe zu berechnen
        textarea.style.height = 'auto';
        // Passe die Höhe an den Inhalt an
        textarea.style.height = `${textarea.scrollHeight}px`;
    }

    handleKeyUp(e) {
        // Überprüfe, ob Strg + Enter gedrückt wurde
        if (e.key === 'Enter' && e.ctrlKey) {
            this.handleSend();
        }
    }

    handleSend() {
        if (!this.processing && this.value.trim() !== '') {
            this.dispatchEvent(new CustomEvent('send-message', {detail: {message: this.value}}));
            this.value = '';
        }
    }
}

customElements.define('chat-input', InputBar);

// ChatApp Component
class ChatApp extends LitElement {

    connectedCallback() {
        super.connectedCallback();
        // this.loadChatContext();
        this.setupStream();
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
        this.isLoading = true;
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
            await this.loadChatContext();
        } catch (error) {
            console.error('Error clearing chat:', error);
        } finally {
            this.isLoading = false;
        }
    }

    static properties = {
        messages: {type: Array},
        isLoading: {},
        eventSource: {type: Object},
    };

    constructor() {
        super();
        this.messages = [];
        this.isLoading = false;
        this.needScoll = false;
        this.eventSource = null;
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

    setupStream() {
        this.eventSource = new EventSource('http://localhost:8080/chat/default/context');
        this.eventSource.onfinish =  () => console.warn("finish")

        this.eventSource.onmessage = (event) => {
            const message = JSON.parse(event.data);

            let lastMessage = this.messages.length - 1;
            const lastRole = this.messages.length > 0 ? this.messages[lastMessage].role : undefined

            if (lastRole !== message.role) {
                console.warn("NEW MESSAGE!!! lastMessage", lastMessage, "lastRole", lastRole, "message.role", message.role, message)
                lastMessage++
                this.messages = [...this.messages, message] ;
            } else {
                console.warn("APPEND MESSAGE!!! lastMessage", lastMessage, "lastRole", lastRole, "message.role", message.role, message)
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
        };
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
                        <button @click=${this.compactChat}>Compact</button>
                        <button @click=${this.clearChat}>Clear</button>
                    </div>
                    <div class="messages">
                        ${this.messages.map(msg => html`
                            <chat-message .message=${msg}></chat-message>`)}
                    </div>
                `}
                <div class="input-area">
                    ${initial ? html`<h3>Was geht up?</h3>` : undefined}
                    <chat-input .processing=${this.isLoading}
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
        this.isLoading = true;
        try {
            // Send message to the REST API
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
        } finally {
            this.isLoading = false;
        }
    }
}

customElements.define('chat-app', ChatApp);