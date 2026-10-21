<script lang="ts">
  import { tr } from '../i18n';
  import type { ConnectionStatus } from '../ws';

  export let status: ConnectionStatus = 'connecting';
  export let compact = false;

  const labels: Record<ConnectionStatus, string> = {
    connecting: 'common.connecting',
    live: 'common.live',
    offline: 'common.reconnecting'
  };
</script>

<span class="connection-dot {status}" title={$tr(labels[status])}>
  {#if !compact}{$tr(labels[status])}{/if}
</span>
