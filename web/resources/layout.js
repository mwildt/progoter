import {css, html, LitElement} from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';
import './atoms.js';


// SidebarLayout Component
class SidebarLayout extends LitElement {
    static styles = css`
        .sidebar-layout-container {
            display: grid;
            grid-template-columns: 200px 1fr;
            height: 100vh;
            width: 100vw;
            margin: 0;
            padding: 0;
        }

        .sidebar {
            width: 200px;
            border-right: 1px solid #e9ecef;
            padding: 10px;
            background-color: #f8f9fa;
            overflow-y: auto;
            max-height: calc(100vh - 20px);
        }

        .main-content {
            flex: 1;
            overflow: hidden;
            width: 100%;
        }
    `;

    render() {
        return html`
            <div class="sidebar-layout-container">
                <div class="sidebar">
                    <slot name="sidebar"></slot>
                </div>
                <div class="main-content">
                    <slot></slot>
                </div>
            </div>
        `;
    }
}

customElements.define('sidebar-layout', SidebarLayout);