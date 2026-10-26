<script lang="ts">
  export let name:
    | 'search'
    | 'plus'
    | 'close'
    | 'chevron'
    | 'shuffle'
    | 'play'
    | 'skip'
    | 'up'
    | 'down'
    | 'lock'
    | 'qr'
    | 'volume'
    | 'check'
    | 'alert'
    | 'microphone'
    | 'link'
    | 'phone' = 'chevron';
  export let size = 24;
  export let strokeWidth = 1.5;
</script>

<svg
  viewBox="0 0 24 24"
  width={size}
  height={size}
  fill="none"
  stroke="currentColor"
  stroke-width={strokeWidth}
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"
  focusable="false"
>
  {#if name === 'search'}
    <circle cx="11" cy="11" r="7" />
    <path d="M16.5 16.5 21 21" />
  {:else if name === 'plus'}
    <path d="M12 5v14M5 12h14" />
  {:else if name === 'close'}
    <path d="M6 6l12 12M18 6 6 18" />
  {:else if name === 'chevron'}
    <path d="M9 5l7 7-7 7" />
  {:else if name === 'shuffle'}
    <path d="M4 7h3l10 10h3M4 17h3l4-4" />
    <path d="M17 4l3 3-3 3M17 14l3 3-3 3" />
  {:else if name === 'play'}
    <path d="M8 5l11 7-11 7z" />
  {:else if name === 'skip'}
    <path d="M6 5l10 7-10 7zM18 5v14" />
  {:else if name === 'up'}
    <path d="M12 19V5M6 11l6-6 6 6" />
  {:else if name === 'down'}
    <path d="M12 5v14M6 13l6 6 6-6" />
  {:else if name === 'lock'}
    <rect x="5" y="11" width="14" height="9" />
    <path d="M8 11V8a4 4 0 0 1 8 0v3" />
  {:else if name === 'qr'}
    <rect x="4" y="4" width="6" height="6" />
    <rect x="14" y="4" width="6" height="6" />
    <rect x="4" y="14" width="6" height="6" />
    <path d="M14 14h3v3M20 20h-3M20 14v2" />
  {:else if name === 'volume'}
    <path d="M4 9v6h4l5 4V5L8 9z" />
    <path d="M17 9a4 4 0 0 1 0 6" />
  {:else if name === 'check'}
    <path d="M5 13l4 4L19 7" />
  {:else if name === 'alert'}
    <path d="M12 4l9 16H3z" />
    <path d="M12 10v4M12 18h.01" />
  {:else if name === 'microphone'}
    <rect x="9" y="3" width="6" height="11" rx="3" />
    <path d="M5 11a7 7 0 0 0 14 0M12 18v3" />
  {:else if name === 'link'}
    <path d="M10 13a5 5 0 0 0 7 0l2-2a5 5 0 0 0-7-7l-1 1" />
    <path d="M14 11a5 5 0 0 0-7 0l-2 2a5 5 0 0 0 7 7l1-1" />
  {:else if name === 'phone'}
    <rect x="7" y="3" width="10" height="18" rx="2" />
    <path d="M11 18h2" />
  {/if}
</svg>
