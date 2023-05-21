import { css, html, LitElement } from "lit";
import { customElement, property } from "lit/decorators.js";
import { createRef, Ref, ref } from "lit/directives/ref.js";

// -- Monaco Editor Imports --
import * as monaco from "monaco-editor";
// @ts-ignore
import styles from "monaco-editor/min/vs/editor/editor.main.css?inline";
// @ts-ignore
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";

// @ts-ignore
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker";

// @ts-ignore
import cssWorker from "monaco-editor/esm/vs/language/css/css.worker?worker";

// @ts-ignore
import htmlWorker from "monaco-editor/esm/vs/language/html/html.worker?worker";

// @ts-ignore
import tsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

// @ts-ignore
self.MonacoEnvironment = {
    getWorker(_: any, label: string) {
        if (label === "json") {
            return new jsonWorker();
        }
        if (label === "css" || label === "scss" || label === "less") {
            return new cssWorker();
        }
        if (label === "html" || label === "handlebars" || label === "razor") {
            return new htmlWorker();
        }
        if (label === "typescript" || label === "javascript") {
            return new tsWorker();
        }
        return new editorWorker();
    },
};

@customElement("spec-editor")
export class SpecEditor extends LitElement {
    private container: Ref<HTMLElement> = createRef();
    editor?: monaco.editor.IStandaloneCodeEditor;
    @property({ type: Boolean, attribute: "readonly" }) readOnly?: boolean;
    @property() theme?: string;
    @property() language?: string;
    @property() code?: string;

    static styles = css`
    :host {
      --editor-width: 100%;
      --editor-height: 100%;
    }
    main {
      width: var(--editor-width);
      height: var(--editor-height);
    }
  `;

    render() {
        return html`
      <style>
        ${styles}
      </style>
      <main ${ref(this.container)}></main>
    `;
    }

    private getFile() {
        if (this.children.length > 0) return this.children[0];
        return null;
    }

    private getCode() {
        if (this.code) return this.code;
        const file = this.getFile();
        if (!file) return;
        return file.innerHTML.trim();
    }

    private getTheme() {
        if (this.theme) return this.theme;
        if (this.isDark()) return "vs-dark";
        return "vs-light";
    }

    private isDark() {
        return (
            window.matchMedia &&
            window.matchMedia("(prefers-color-scheme: dark)").matches
        );
    }

    setValue(value: string) {
        this.editor!.setValue(value);
    }

    getValue() {
        const value = this.editor!.getValue();
        return value;
    }


    firstUpdated() {

        const options = {
            base: 'vs-dark',
            renderSideBySide: false,
            inherit: true,
            rules: [

                {
                    "foreground": "E400FB",
                    "token": "string"
                },
                {
                    "foreground": "62C4FFFF",
                    "token": "type"
                },
            ],
            colors: {
                'editor.foreground': '#ffffff',
                'editor.background': '#0d1117',
                'editorCursor.foreground': '#62C4FFFF',
                'editor.lineHighlightBackground': '#E400FB30',
                'editorLineNumber.foreground': '#6368747F',
                'editorLineNumber.activeForeground': '#E400FB',
                'editor.inactiveSelectionBackground': '#FF3C742D',
                'diffEditor.removedTextBackground': '#FF3C741A',
                'diffEditor.insertedTextBackground': '#62C4FF1A',
            }
        };
        // @ts-ignore
        monaco.editor.defineTheme("pb33f", options);
        monaco.editor.setTheme('pb33f');


        this.editor = monaco.editor.create(this.container.value!, {
            value: this.getCode(),
            language: 'yaml',
            theme: 'pb33f',
            automaticLayout: true,
            readOnly: true,
            minimap: {
                enabled: false
            },
            scrollbar: {
                vertical: 'hidden',
            }
        });
        this.editor.getModel()!.onDidChangeContent(() => {
            this.dispatchEvent(new CustomEvent("change", { detail: {} }));
        });
        window
            .matchMedia("(prefers-color-scheme: dark)")
            .addEventListener("change", () => {
                monaco.editor.setTheme(this.getTheme());
            });
    }
}

declare global {
    interface HTMLElementTagNameMap {
        "spec-editor": SpecEditor;
    }
}