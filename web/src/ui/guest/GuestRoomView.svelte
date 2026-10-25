<script lang="ts">
  import { onMount } from 'svelte';
  import { tr } from '../../application/i18n';
  import { createGuestController } from '../../application/guest/controller';
  import { getAppContext } from '../../ports/context';
  import { computeElapsed, secondsUntil } from '../../domain/playback';
  import GuestHeader from './GuestHeader.svelte';
  import GuestHostBar from './GuestHostBar.svelte';
  import GuestModals from './GuestModals.svelte';
  import GuestNowPlayingCard from './GuestNowPlayingCard.svelte';
  import GuestQueueList from './GuestQueueList.svelte';
  import GuestSearchTab from './GuestSearchTab.svelte';

  export let code: string;

  const c = createGuestController(code, getAppContext());
  const {
    snap, status, ended, toast, loadError, tab, q, results, searching, searched, stale,
    suggestions, loadingSuggestions,
    searchFailed, selected, nickname, nicknameError, adding, addError, pasteUrl, pasteError,
    pinSheet, pin, pinError, unlocking, manageOpen, confirmClear, confirmEnd, busy,
    libraryCount, libraryBusy, libraryError,
    load, queue: queueSearch, clear: clearSearch, shuffle: shuffleSuggestions, choose,
    preparePaste, submit,
    toggleVote, startSkip, castSkip, cancelSkip, removeSong, playback, moveSong, unlock,
    toggleLock, clearQueue, loadLibrary, importLibrary, endRoom, newRoom
  } = c;

  let nowTick = Date.now();

  onMount(() => {
    const ticker = setInterval(() => (nowTick = Date.now()), 1000);
    load();
    return () => clearInterval(ticker);
  });

  $: current = $snap;
  $: voteLeft = current?.skipVote?.expiresAt ? secondsUntil(current.skipVote.expiresAt, nowTick) : 0;
  $: cooldownLeft = current?.skipVote?.cooldownUntil ? secondsUntil(current.skipVote.cooldownUntil, nowTick) : 0;
  $: elapsedNow = current ? computeElapsed(current, nowTick) : 0;
  $: queue = current?.queue ?? [];
  $: locked = current?.room.queueLocked ?? false;
  $: isHost = current?.you.hostUnlocked ?? false;
  $: full = queue.length >= (current?.room.queueCapacity ?? Number.MAX_SAFE_INTEGER);
</script>

<svelte:head>
  <title>{current ? `${current.room.code.toUpperCase()} · Fana Karaoke` : 'Fana Karaoke'}</title>
</svelte:head>

{#if $ended}
  <main class="center-page">
    <div class="card end-card">
      <span class="badge badge-hot">{$tr('app.name')}</span>
      <h1>{$tr('tv.endedTitle')}</h1>
      <p class="muted">{$tr('tv.endedBody')}</p>
      <button class="btn btn-primary" type="button" on:click={newRoom}>{$tr('common.createRoom')}</button>
      <a class="btn btn-ghost" href="/">{$tr('common.home')}</a>
    </div>
  </main>
{:else if $loadError}
  <main class="center-page">
    <div class="card end-card">
      <span class="badge">Fana</span>
      <h1>{$tr(`error.${$loadError}`)}</h1>
      <a class="btn btn-primary" href="/">{$tr('common.home')}</a>
    </div>
  </main>
{:else if !current}
  <main class="center-page">
    <span class="spinner large"></span>
    <p class="muted">{$tr('common.loading')}</p>
  </main>
{:else}
  <div class="guest" class:host={isHost}>
    <GuestHeader
      room={current.room}
      status={$status}
      {isHost}
      tab={$tab}
      queueCount={queue.length}
      onUnlock={() => pinSheet.set(true)}
      onTab={(value) => tab.set(value)}
    />

    <main class="guest-main">
      {#if $tab === 'search'}
        <GuestSearchTab
          {q}
          {results}
          {suggestions}
          {loadingSuggestions}
          {searching}
          {searched}
          {stale}
          {searchFailed}
          {pasteUrl}
          {pasteError}
          {locked}
          {full}
          onSearch={queueSearch}
          onClear={clearSearch}
          onChoose={choose}
          onShuffle={shuffleSuggestions}
          onPaste={preparePaste}
        />
      {:else}
        <section class="queue-panel">
          <GuestNowPlayingCard
            snap={current}
            {isHost}
            {voteLeft}
            {cooldownLeft}
            {elapsedNow}
            onStartSkip={startSkip}
            onCastSkip={castSkip}
            onCancelSkip={cancelSkip}
            onSearch={() => tab.set('search')}
          />
          <GuestQueueList
            {queue}
            {isHost}
            onVote={toggleVote}
            onMove={moveSong}
            onRemove={removeSong}
          />
        </section>
      {/if}
    </main>

    {#if isHost}
      <GuestHostBar
        snap={current}
        busy={$busy}
        onPlayback={playback}
        onToggleLock={toggleLock}
        onManage={() => {
          manageOpen.set(true);
          loadLibrary();
        }}
      />
    {/if}

    {#if $toast}
      <div class="toast" role="status" aria-live="polite">{$toast}</div>
    {/if}

    <GuestModals
      {selected}
      {nickname}
      {nicknameError}
      {adding}
      {addError}
      {pinSheet}
      {pin}
      {pinError}
      {unlocking}
      {manageOpen}
      {confirmClear}
      {confirmEnd}
      {libraryCount}
      {libraryBusy}
      {libraryError}
      {code}
      {locked}
      onSubmit={submit}
      onUnlock={unlock}
      onToggleLock={toggleLock}
      onClear={clearQueue}
      onImportLibrary={importLibrary}
      onEnd={endRoom}
    />
  </div>
{/if}

<style>
  .guest {
    min-height: 100dvh;
    max-width: 720px;
    margin: 0 auto;
    padding-bottom: 2rem;
  }

  .guest.host {
    padding-bottom: 5.5rem;
  }

  .guest-main {
    padding: 1rem;
  }

  .queue-panel {
    display: grid;
    gap: 1rem;
  }

  @media (min-width: 680px) {
    .guest-main {
      padding-left: 1.5rem;
      padding-right: 1.5rem;
    }
  }
</style>
