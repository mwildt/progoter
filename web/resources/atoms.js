import {css, html, LitElement} from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';

import {marked} from "https://cdn.jsdelivr.net/npm/marked/lib/marked.esm.js";
import DOMPurify from "https://cdn.jsdelivr.net/npm/dompurify@3.0.8/dist/purify.es.mjs";

// Label Component
class AtomicLabel extends LitElement {
    static properties = {
        for: {type: String},
    };

    constructor() {
        super();
        this.for = '';
    }

    static styles = css`
        label {
            display: block;
            margin-bottom: 5px;
            font-size: 13px;
            color: #495057;
            font-weight: 500;
        }
    `;

    render() {
        return html`
            <label for=${this.for}>
                <slot></slot>
            </label>
        `;
    }
}

customElements.define('atomic-label', AtomicLabel);

// Input Component
class AtomicInput extends LitElement {
    static properties = {
        type: {type: String},
        value: {type: String},
        placeholder: {type: String},
        required: {type: Boolean},
    };

    constructor() {
        super();
        this.type = 'text';
        this.value = '';
        this.placeholder = '';
        this.required = false;
    }

    static styles = css`
        input {
            width: 100%;
            padding: 8px 10px;
            margin-bottom: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 13px;
            transition: border-color 0.2s ease;
            box-sizing: border-box;
        }

        input:focus {
            outline: none;
            border-color: #007bff;
        }
    `;

    render() {
        return html`
            <input
                type=${this.type}
                .value=${this.value}
                placeholder=${this.placeholder}
                ?required=${this.required}
                @input=${this.handleInput}
            >
        `;
    }

    handleInput(e) {
        this.value = e.target.value;
        this.dispatchEvent(new CustomEvent('input-change', {detail: {value: this.value}}));
    }
}

customElements.define('atomic-input', AtomicInput);

// Form Actions Component
class AtomicFormActions extends LitElement {
    static styles = css`
        .form-actions {
            display: flex;
            gap: 10px;
            margin-top: 10px;
        }

        .primary {
            background-color: #28a745;
            color: white;
            border: 1px solid #28a745;
        }

        .primary:hover {
            background-color: #218838;
        }

        .secondary {
            background-color: transparent;
            color: #6c757d;
            border: 1px solid #6c757d;
        }

        .secondary:hover {
            background-color: rgba(108, 117, 125, 0.1);
        }
    `;

    render() {
        return html`
            <div class="form-actions">
                <slot></slot>
            </div>
        `;
    }
}

customElements.define('atomic-form-actions', AtomicFormActions);

// Button Component
class AtomicButton extends LitElement {
    static properties = {
        label: {type: String},
        disabled: {type: Boolean},
        variant: {type: String},
    };

    constructor() {
        super();
        this.label = '';
        this.disabled = false;
        this.variant = 'primary';
    }

    static styles = css`
        :host {
           display: flex;
        }
    
        button {
            padding: 6px 12px;
            background-color: transparent;
            border: 1px solid;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            transition: all 0.2s ease;
        }

        button.primary {
            color: #28a745;
            border-color: #28a745;
        }

        button.primary:hover {
            background-color: rgba(40, 167, 69, 0.1);
        }

        button.secondary {
            color: #6c757d;
            border-color: #6c757d;
        }

        button.secondary:hover {
            background-color: rgba(108, 117, 125, 0.1);
        }

        button:disabled {
            color: #cccccc;
            border-color: #cccccc;
            cursor: not-allowed;
        }
    `;

    render() {
        return html`
            <button class=${this.variant} ?disabled=${this.disabled} @click=${this.handleClick}>
                ${this.label}
            </button>
        `;
    }

    handleClick() {
        this.dispatchEvent(new CustomEvent('button-click'));
    }
}

customElements.define('atomic-button', AtomicButton);

// MarkdownView Component
class AtomicMarkdown extends LitElement {
    static properties = {
        content: {type: String},
    };

    constructor() {
        super();
        this.content = "# Hallo\n**Markdown sicher gerendert**";
    }

    static styles = css`
    .md {
      line-height: 1.6;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
      color: #333;
      font-size: 14px;
    }

    .md pre {
      padding: 12px;
      background-color: #f8f9fa;
      border-radius: 4px;
      overflow: auto;
      border-left: 3px solid #007bff;
    }

    .md code {
      font-family: 'Courier New', monospace;
      font-size: 13px;
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

customElements.define("atomic-markdown", AtomicMarkdown);