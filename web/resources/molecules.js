// NewContextForm Component
class NewContextForm extends LitElement {
    static properties = {
        visible: {type: Boolean},
        contextId: {type: String},
        basePath: {type: String}
    };

    constructor() {
        super();
        this.visible = false;
        this.contextId = '';
        this.basePath = '';
    }

    static styles = css`
        .new-context-form {
            padding: 20px;
            background-color: #fff;
            border-radius: 4px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
        }

        .new-context-form h3 {
            margin-top: 0;
            margin-bottom: 15px;
            font-size: 16px;
            color: #495057;
            font-weight: 500;
        }
    `;

    async handleSubmit(e) {
        e.preventDefault();
        const contextId = this.contextId;
        const basePath = this.basePath;
        
        if (contextId) {
            const response = await fetch('/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({id: contextId, basePath: basePath || ''})
            }).then(r => r.json())
                .then(context => {
                    this.dispatchEvent(new CustomEvent('context-created', {
                        detail: context,
                        bubbles: true,
                        composed: true
                    }));
                }).catch(error => console.error('Failed to create context', error))
        }
    }

    handleCancel() {
        this.dispatchEvent(new CustomEvent('cancel', {
            bubbles: true,
            composed: true
        }));
    }

    handleContextIdChange(e) {
        this.contextId = e.detail.value;
    }

    handleBasePathChange(e) {
        this.basePath = e.detail.value;
    }

    render() {
        if (!this.visible) return html``;
        
        return html`
            <div class="new-context-form">
                <h3>Create New Context</h3>
                <form @submit=${this.handleSubmit}>
                    <atomic-label for="contextId">Context ID:</atomic-label>
                    <atomic-input
                        id="contextId"
                        .value=${this.contextId}
                        placeholder="Enter context ID"
                        required
                        @input-change=${this.handleContextIdChange}
                    ></atomic-input>
                    <atomic-label for="basePath">Base Path (optional):</atomic-label>
                    <atomic-input
                        id="basePath"
                        .value=${this.basePath}
                        placeholder="Enter base path"
                        @input-change=${this.handleBasePathChange}
                    ></atomic-input>
                    <atomic-form-actions>
                        <atomic-button variant="primary" type="submit" label="Create" @button-click=${this.handleSubmit}></atomic-button>
                        <atomic-button variant="secondary" type="button" label="Cancel" @button-click=${this.handleCancel}></atomic-button>
                    </atomic-form-actions>
                </form>
            </div>
        `;
    }
}

customElements.define('new-context-form', NewContextForm);


// ContextList Component
class ContextList extends LitElement {
    static properties = {
        contexts: {type: Array},
        selectedContext: {type: String},
        showNewContextForm: {type: Boolean}
    };

    constructor() {
        super();
        this.contexts = [];
        this.selectedContext = undefined;
        this.showNewContextForm = false;
    }

    connectedCallback() {
        super.connectedCallback();
        this.fetchContexts();
    }

    disconnectedCallback() {
        super.disconnectedCallback();
    }

    handleContextDeleted(e) {
        this.contexts = this.contexts.filter(context => context.id !== e.detail.contextId);
        if (this.contexts.length > 0) {
            this.selectedContext = this.contexts[0];
        } else {
            this.selectedContext = null;
            this.showNewContextForm = true;
        }
    }

    async handleContextCreated(e) {
        await this.fetchContexts()
        this.selectedContext = e.detail;
        this.showNewContextForm = false;
    }

    handleCancelNewContext() {
        this.showNewContextForm = false;
    }

    static styles = css`
        .context-list-container {
            display: grid;
            grid-template-columns: 220px 1fr;
            height: 100vh;
            width: 100vw;
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
        }

        .context-list {
            border-right: 1px solid #e9ecef;
            padding: 15px;
            background-color: #f8f9fa;
            overflow-y: auto;
            max-height: calc(100vh - 20px);
        }

        .context-list h3 {
            margin-top: 0;
            margin-bottom: 15px;
            color: #495057;
            font-size: 16px;
            font-weight: 500;
            letter-spacing: 0.5px;
        }

        .context-item {
            padding: 10px 12px;
            margin-bottom: 6px;
            background-color: transparent;
            border-radius: 4px;
            cursor: pointer;
            transition: background-color 0.2s ease;
            display: flex;
            align-items: center;
            color: #495057;
        }

        .context-item:hover {
            background-color: rgba(0, 0, 0, 0.03);
        }

        .context-item.selected {
            background-color: rgba(0, 123, 255, 0.1);
            color: #007bff;
            font-weight: 500;
        }

        .context-item.selected:hover {
            background-color: rgba(0, 123, 255, 0.15);
        }

        .context-item-icon {
            margin-right: 8px;
            font-size: 14px;
            color: #6c757d;
        }

        .context-item.selected .context-item-icon {
            color: #007bff;
        }

        .new-context-button {
            margin-top: 15px;
            padding: 8px;
            background-color: transparent;
            color: #28a745;
            border: 1px solid #28a745;
            border-radius: 4px;
            cursor: pointer;
            width: 100%;
            font-weight: 500;
            transition: all 0.2s ease;
            font-size: 14px;
        }

        .new-context-button:hover {
            background-color: rgba(40, 167, 69, 0.1);
        }

        .chat-app-container {
            flex: 1;
            overflow: hidden;
            width: 100%;
            background-color: #fff;
        }

        .empty-state {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100%;
            text-align: center;
            color: #6c757d;
            font-size: 14px;
        }
    `;

    async fetchContexts() {
        try {
            const response = await fetch('/chat');
            if (response.ok) {
                this.contexts = await response.json();
                if (this.contexts.length > 0 && !this.selectedContext) {
                    this.selectedContext = this.contexts[0];
                } else if (this.contexts.length === 0) {
                    this.showNewContextForm = true;
                }
            } else {
                console.error('Failed to fetch contexts');
            }
        } catch (error) {
            console.error('Error fetching contexts:', error);
        }
    }

    handleContextClick(context) {
        this.selectedContext = context;
        this.showNewContextForm = false;
    }

    handleNewContext() {
        this.showNewContextForm = true;
    }

    render() {
        return html`
            <div class="context-list-container">
                <div class="context-list">
                    <h3>Contexts</h3>
                    <div>
                        ${this.contexts.map(context => html`
                            <div
                                    class="context-item ${this.selectedContext.id === context.id ? 'selected' : ''}"
                                    @click=${() => this.handleContextClick(context)}
                            >
                                <span class="context-item-icon">📄</span>
                                <span>${context.id}</span>
                                ${context.basePath ? html`<span class="context-base-path"> (${context.basePath})</span>` : ''}
                            </div>
                        `)}
                    </div>
                    <button class="new-context-button" @click=${this.handleNewContext}>+ New Context</button>
                </div>
                <div class="chat-app-container">
                    ${this.showNewContextForm || this.contexts.length === 0 ? html`
                        <new-context-form .visible=${true}
                            .contextId=${this.contexts.length === 0 ? 'default' : ''}
                            .basePath=${this.contexts.length === 0 ? '.' : ''}
                            @context-created=${this.handleContextCreated}
                            @cancel=${this.handleCancelNewContext}
                        ></new-context-form>
                    ` : html`
                        ${this.selectedContext ? html`
                            <chat-app .context=${this.selectedContext}
                                @context-deleted=${this.handleContextDeleted}
                            ></chat-app>
                        ` : html`
                            <div class="empty-state">
                                <p>No context selected. Please create a new context or select an existing one.</p>
                            </div>
                        `}
                    `}
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