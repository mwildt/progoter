import {css, html, LitElement} from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';

import {marked} from "https://cdn.jsdelivr.net/npm/marked/lib/marked.esm.js";
import DOMPurify from "https://cdn.jsdelivr.net/npm/dompurify@3.0.8/dist/purify.es.mjs";

// Button Component
class AtomicButton extends LitElement {
    static properties = {
        label: {type: String},
        disabled: {type: Boolean},
    };

    constructor() {
        super();
        this.label = '';
        this.disabled = false;
    }

    static styles = css`
        :host {
           display: flex;
        }
    
        button {
            padding: 8px 16px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
        }

        button:hover {
            background-color: #0056b3;
        }

        button:disabled {
            background-color: #cccccc;
            cursor: not-allowed;
        }
    `;

    render() {
        return html`
            <button ?disabled=${this.disabled} @click=${this.handleClick}>
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

customElements.define("atomic-markdown", AtomicMarkdown);