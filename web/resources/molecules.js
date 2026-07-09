// ContextList Component
class ContextList extends LitElement {
    static properties = {
        contexts: {type: Array},
        selectedContext: {type: String}
    };

    constructor() {
        super();
        this.contexts = [];
        this.selectedContext = 'default';
    }

    connectedCallback() {
        super.connectedCallback();
        this.fetchContexts();
    }

    static styles = css`
        .context-list-container {
            display: grid;
            grid-template-columns: 200px 1fr;
            height: 100vh;
            width: 100vw;
            margin: 0;
            padding: 0;
        }

        .context-list {
            border-right: 1px solid #e9ecef;
            padding: 10px;
            background-color: #f8f9fa;
            overflow-y: auto;
            max-height: calc(100vh - 20px);
        }

        .context-item {
            padding: 10px;
            margin-bottom: 5px;
            background-color: #fff;
            border-radius: 4px;
            cursor: pointer;
            transition: background-color 0.2s;
        }

        .context-item:hover {
            background-color: #e9ecef;
        }

        .context-item.selected {
            background-color: #007bff;
            color: white;
            font-weight: bold;
        }

        .context-item.selected:hover {
            background-color: #0069d9;
        }

        .new-context-button {
            margin-top: 10px;
            padding: 10px;
            background-color: #28a745;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            width: 100%;
        }

        .new-context-button:hover {
            background-color: #218838;
        }

        .chat-app-container {
            flex: 1;
            overflow: hidden;
            width: 100%;
        }
    `;

    async fetchContexts() {
        try {
            const response = await fetch('/chat');
            if (response.ok) {
                this.contexts = await response.json();
                if (this.contexts.length > 0 && !this.selectedContext) {
                    this.selectedContext = this.contexts[0];
                }
            } else {
                console.error('Failed to fetch contexts');
            }
        } catch (error) {
            console.error('Error fetching contexts:', error);
        }
    }

    handleContextClick(contextId) {
        this.selectedContext = contextId;
    }

    handleNewContext() {
        const contextName = prompt('Enter a name for the new context:');
        if (contextName) {
            const basePath = prompt('Enter the base path for the new context (optional):');
            this.createContext(contextName, basePath);
        }
    }

    async createContext(contextId, basePath) {
        try {
            const response = await fetch('/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({id: contextId, basePath: basePath || ''})
            });
            if (response.ok) {
                this.fetchContexts();
            } else {
                console.error('Failed to create context');
            }
        } catch (error) {
            console.error('Error creating context:', error);
        }
    }

    render() {
        return html`
            <div class="context-list-container">
                <div class="context-list">
                    <h3>Contexts</h3>
                    <div>
                        ${this.contexts.map(contextId => html`
                            <div
                                    class="context-item ${this.selectedContext === contextId ? 'selected' : ''}"
                                    @click=${() => this.handleContextClick(contextId)}
                            >
                                ${contextId}
                            </div>
                        `)}
                    </div>
                    <button class="new-context-button" @click=${this.handleNewContext}>New Context</button>
                </div>
                <div class="chat-app-container">
                    <chat-app .contextId=${this.selectedContext}></chat-app>
                </div>
            </div>
        `;
    }
}

customElements.define('context-list', ContextList);


import {css, html, LitElement} from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';
import './atoms.js';


// AppLayout Component
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
                    <atomic-markdown .content=${this.message.content}></atomic-markdown>
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
                <atomic-button label="Send" ?disabled=${this.processing} @button-click=${this.handleSend}></atomic-button>
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