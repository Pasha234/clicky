<x-filament-panels::page>
    <style>
        .site-snippet { display: grid; gap: 1.5rem; }
        .site-snippet__intro { color: #6b7280; line-height: 1.6; }
        .site-snippet__steps { display: grid; gap: .75rem; margin-top: 1.25rem; }
        .site-snippet__step { display: flex; align-items: center; gap: .75rem; color: #4b5563; font-size: .875rem; }
        .site-snippet__number { display: grid; place-items: center; width: 1.5rem; height: 1.5rem; flex: 0 0 auto; border-radius: 50%; background: #eef2ff; color: #4f46e5; font-size: .75rem; font-weight: 700; }
        .site-snippet__block { margin-top: 1.5rem; }
        .site-snippet__block-header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; margin-bottom: .65rem; }
        .site-snippet__label { color: #374151; font-size: .875rem; font-weight: 700; }
        .site-snippet__token { display: block; overflow-wrap: anywhere; border-radius: .75rem; padding: 1rem; background: #111827; color: #e5e7eb; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .8125rem; line-height: 1.5; }
        .site-snippet__code { overflow-x: auto; margin: 0; border-radius: .75rem; padding: 1.25rem; background: #111827; color: #e5e7eb; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .8125rem; line-height: 1.65; tab-size: 2; }
        .site-snippet__copy { display: inline-flex; align-items: center; justify-content: center; min-width: 5.5rem; border: 1px solid #d1d5db; border-radius: .6rem; padding: .45rem .75rem; cursor: pointer; background: #fff; color: #374151; font-size: .8125rem; font-weight: 700; transition: background .15s, border-color .15s, color .15s; }
        .site-snippet__copy:hover { border-color: #818cf8; background: #eef2ff; color: #4338ca; }
        .site-snippet__copy.is-copied { border-color: #6ee7b7; background: #ecfdf5; color: #047857; }
        .dark .site-snippet__intro, .dark .site-snippet__step { color: #9ca3af; }
        .dark .site-snippet__label { color: #e5e7eb; }
        .dark .site-snippet__number { background: rgba(99, 102, 241, .2); color: #a5b4fc; }
        .dark .site-snippet__copy { border-color: rgba(255, 255, 255, .15); background: rgba(255, 255, 255, .04); color: #e5e7eb; }
        .dark .site-snippet__copy:hover { border-color: #818cf8; background: rgba(99, 102, 241, .15); color: #c7d2fe; }
        .dark .site-snippet__copy.is-copied { border-color: rgba(52, 211, 153, .4); background: rgba(16, 185, 129, .12); color: #6ee7b7; }
    </style>

    <div class="site-snippet" x-data="{
        copied: false,
        async copySnippet() {
            const text = this.$refs.snippet.textContent.trim()

            try {
                await navigator.clipboard.writeText(text)
            } catch (error) {
                const input = document.createElement('textarea')
                input.value = text
                input.style.position = 'fixed'
                input.style.opacity = '0'
                document.body.appendChild(input)
                input.select()
                document.execCommand('copy')
                input.remove()
            }

            this.copied = true
            window.setTimeout(() => this.copied = false, 2000)
        },
    }">
        <x-filament::section>
            <x-slot name="heading">Install tracking for {{ $site->name }}</x-slot>
            <x-slot name="description">Copy the code below and add it once to your site.</x-slot>

            <p class="site-snippet__intro">Place the snippet just before the closing <code>&lt;/body&gt;</code> tag. It records click events from every page where it is installed.</p>

            <div class="site-snippet__steps">
                <div class="site-snippet__step"><span class="site-snippet__number">1</span>Copy the tracking snippet.</div>
                <div class="site-snippet__step"><span class="site-snippet__number">2</span>Paste it into your site's shared layout or footer.</div>
                <div class="site-snippet__step"><span class="site-snippet__number">3</span>Visit your site and check Analytics after a moment.</div>
            </div>
        </x-filament::section>

        <x-filament::section>
            <x-slot name="heading">Tracking code</x-slot>
            <x-slot name="description">This snippet is configured specifically for this site.</x-slot>

            <div class="site-snippet__block">
                <div class="site-snippet__block-header">
                    <span class="site-snippet__label">Snippet</span>
                    <button type="button" class="site-snippet__copy" x-on:click="copySnippet()" x-bind:class="{ 'is-copied': copied }" x-bind:aria-label="copied ? 'Snippet copied' : 'Copy tracking snippet'">
                        <span x-text="copied ? 'Copied!' : 'Copy code'"></span>
                    </button>
                </div>
                <pre class="site-snippet__code"><code x-ref="snippet">{{ $snippet }}</code></pre>
            </div>

            <div class="site-snippet__block">
                <div class="site-snippet__block-header">
                    <span class="site-snippet__label">Tracking token</span>
                </div>
                <code class="site-snippet__token">{{ $token ?? 'No active token' }}</code>
            </div>
        </x-filament::section>
    </div>
</x-filament-panels::page>
