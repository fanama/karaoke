<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { api, ApiError } from '../lib/api';
  import ConnectionDot from '../lib/components/ConnectionDot.svelte';
  import LanguageToggle from '../lib/components/LanguageToggle.svelte';
  import Modal from '../lib/components/Modal.svelte';
  import { errorText, noticeText, tr } from '../lib/i18n';
  import { formatDuration, secondsUntil } from '../lib/format';
  import type { Notice, SearchResult, Snapshot } from '../lib/types';
  import { connectRoom, type ConnectionStatus, type RoomSocket } from '../lib/ws';

  export let code: string;

  let snap: Snapshot | null = null;
  let status: ConnectionStatus = 'connecting';
  let loadError = '';
  let ended = false;
  let socket: RoomSocket | null = null;
  let tab: 'search' | 'queue' = 'search';
  let nowTick = Date.now();

  let q = '';
  let results: SearchResult[] = [];
  let searching = false;
  let searched = false;
  let stale = false;
  let searchFailed = false;
  let searchSeq = 0;
  let searchTimer: ReturnType<typeof setTimeout> | null = null;

  let selected: { videoId?: string; url?: string; title?: string; channel?: string; thumbnailUrl?: string } | null = null;
  let nickname = '';
  let nicknameError = '';
  let adding = false;
  let addError = '';
  let pasteUrl = '';
  let pasteError = '';

  let pinSheet = false;
  let pin = '';
  let pinError = '';
  let unlocking = false;
  let manageOpen = false;
  let confirmClear = false;
  let confirmEnd = false;
  let busy = false;

  let toast = '';
  let toastTimer: ReturnType<typeof setTimeout> | null = null;
  let roomCheckTimer: ReturnType<typeof setTimeout> | null = null;
  let lastNoticeId = '';
  let seenFirstSnapshot = false;

  let voteLeft = 0;
  let cooldownLeft = 0;
  let elapsedNow = 0;

  $: if (snap) {
    const _tick = nowTick;
    voteLeft = snap.skipVote?.expiresAt ? secondsUntil(snap.skipVote.expiresAt) : 0;
    cooldownLeft = snap.skipVote?.cooldownUntil ? secondsUntil(snap.skipVote.cooldownUntil) : 0;
    const np = snap.nowPlaying;
    elapsedNow = np
      ? np.elapsedSeconds +
        (np.status === 'playing'
          ? Math.max(0, (Date.now() - new Date(snap.serverTime).getTime()) / 1000)
          : 0)
      : 0;
  }

  $: queue = snap?.queue ?? [];
  $: locked = snap?.room.queueLocked ?? false;
  $: isHost = snap?.you.hostUnlocked ?? false;

  onMount(() => {
    let ticker = setInterval(() => (nowTick = Date.now()), 1000);
    load();
    return () => {
      clearInterval(ticker);
      socket?.close();
      if (toastTimer) clearTimeout(toastTimer);
      if (searchTimer) clearTimeout(searchTimer);
      if (roomCheckTimer) clearTimeout(roomCheckTimer);
    };
  });

  async function load() {
    try {
      snap = await api.getRoom(code);
      initialNotice(snap);
      connect();
    } catch (err) {
      loadError = err instanceof ApiError ? err.code : 'connection';
    }
  }

  function initialNotice(s: Snapshot) {
    lastNoticeId = s.notices[0]?.id ?? '';
    seenFirstSnapshot = true;
  }

  function connect() {
    socket?.close();
    socket = connectRoom(code, 'phone', {
      onStatus: (s) => {
        status = s;
        if (s === 'offline') scheduleRoomCheck();
        else if (roomCheckTimer) {
          clearTimeout(roomCheckTimer);
          roomCheckTimer = null;
        }
      },
      onSnapshot: (s) => {
        snap = s;
        handleNotice(s);
      },
      onEnded: () => {
        ended = true;
        socket?.close();
      }
    });
  }

  function scheduleRoomCheck() {
    if (roomCheckTimer) return;
    roomCheckTimer = setTimeout(async () => {
      roomCheckTimer = null;
      try {
        await api.getRoom(code);
      } catch (err) {
        if (err instanceof ApiError && (err.code === 'room_not_found' || err.status === 404)) {
          ended = true;
          socket?.close();
        }
      }
    }, 8000);
  }

  function handleNotice(s: Snapshot) {
    const latest = s.notices[0];
    if (!seenFirstSnapshot || !latest || latest.id === lastNoticeId) return;
    lastNoticeId = latest.id;
    showToast(noticeText($tr, latest.code, latest.params));
  }

  function showToast(text: string) {
    if (!text) return;
    toast = text;
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(() => (toast = ''), 4000);
  }

  function message(err: unknown, fallback = 'error.generic') {
    return err instanceof ApiError ? errorText($tr, err.code) : $tr(fallback);
  }

  function queueSearch() {
    if (searchTimer) clearTimeout(searchTimer);
    searchTimer = setTimeout(runSearch, 350);
  }

  async function runSearch() {
    const value = q.trim();
    if (value.length < 2) {
      results = [];
      searched = false;
      searchFailed = false;
      searching = false;
      return;
    }
    const seq = ++searchSeq;
    searching = true;
    searchFailed = false;
    try {
      const res = await api.search(value);
      if (seq !== searchSeq) return;
      results = res.results;
      stale = res.stale;
      searched = true;
    } catch {
      if (seq !== searchSeq) return;
      results = [];
      searched = true;
      searchFailed = true;
    } finally {
      if (seq === searchSeq) searching = false;
    }
  }

  function clearSearch() {
    q = '';
    results = [];
    searched = false;
    searchFailed = false;
    stale = false;
    if (searchTimer) clearTimeout(searchTimer);
  }

  function chooseResult(result: SearchResult) {
    selected = {
      videoId: result.videoId,
      title: result.title,
      channel: result.channel,
      thumbnailUrl: result.thumbnailUrl
    };
    nickname = '';
    nicknameError = '';
    addError = '';
  }

  function preparePaste() {
    pasteError = '';
    if (!pasteUrl.trim()) {
      pasteError = $tr('error.invalid_url');
      return;
    }
    selected = { url: pasteUrl.trim(), title: pasteUrl.trim() };
    nickname = '';
    nicknameError = '';
    addError = '';
  }

  async function submitAdd() {
    if (!selected || adding) return;
    const name = nickname.trim();
    if (name.length < 1 || name.length > 24) {
      nicknameError = $tr('room.nicknameError');
      return;
    }
    adding = true;
    addError = '';
    try {
      const payload: { videoId?: string; url?: string; nickname: string } = { nickname: name };
      if (selected.videoId) payload.videoId = selected.videoId;
      if (selected.url) payload.url = selected.url;
      snap = await api.addSong(code, payload);
      selected = null;
      nickname = '';
      pasteUrl = '';
      tab = 'queue';
      showToast($tr('room.added'));
    } catch (err) {
      addError = message(err, 'error.video_unavailable');
    } finally {
      adding = false;
    }
  }

  async function toggleVote(itemId: string) {
    try {
      snap = await api.toggleVote(code, itemId);
    } catch (err) {
      showToast(message(err));
    }
  }

  async function startSkip() {
    try {
      snap = await api.startSkipVote(code);
    } catch (err) {
      showToast(message(err));
    }
  }

  async function castSkip() {
    try {
      snap = await api.castSkipVote(code);
    } catch (err) {
      showToast(message(err));
    }
  }

  async function cancelSkip() {
    try {
      snap = await api.cancelSkipVote(code);
    } catch (err) {
      showToast(message(err));
    }
  }

  async function unlock() {
    if (unlocking) return;
    unlocking = true;
    pinError = '';
    try {
      snap = await api.unlockHost(code, pin.trim());
      pinSheet = false;
      pin = '';
      connect();
      showToast($tr('notice.hostUnlocked'));
    } catch (err) {
      pinError = message(err, 'room.wrongPin');
    } finally {
      unlocking = false;
    }
  }

  async function playback(action: 'play' | 'pause' | 'next' | 'cancelCountdown') {
    if (busy) return;
    busy = true;
    try {
      snap = await api.playback(code, action);
    } catch (err) {
      showToast(message(err));
    } finally {
      busy = false;
    }
  }

  async function removeSong(id: string) {
    try {
      snap = await api.removeSong(code, id);
    } catch (err) {
      showToast(message(err));
    }
  }

  async function moveSong(index: number, direction: -1 | 1) {
    if (!snap) return;
    const ids = snap.queue.map((e) => e.id);
    const target = index + direction;
    if (target < 0 || target >= ids.length) return;
    [ids[index], ids[target]] = [ids[target], ids[index]];
    try {
      snap = await api.reorder(code, ids);
    } catch (err) {
      showToast(message(err));
    }
  }

  async function clearQueue() {
    try {
      snap = await api.clearQueue(code);
      confirmClear = false;
      manageOpen = false;
    } catch (err) {
      showToast(message(err));
    }
  }

  async function toggleLock() {
    if (!snap) return;
    try {
      snap = await api.lockQueue(code, !snap.room.queueLocked);
    } catch (err) {
      showToast(message(err));
    }
  }

  async function endRoom() {
    try {
      await api.endRoom(code);
      ended = true;
      manageOpen = false;
      socket?.close();
    } catch (err) {
      showToast(message(err));
    }
  }

  async function newRoom() {
    try {
      const room = await api.createRoom();
      location.assign(room.tvUrl);
    } catch (err) {
      showToast(message(err, 'welcome.createError'));
    }
  }

  function focusOnMount(node: HTMLInputElement) {
    const timer = setTimeout(() => node.focus(), 60);
    return {
      destroy() {
        clearTimeout(timer);
      }
    };
  }
</script>

<svelte:head>
  <title>{snap ? `${snap.room.code.toUpperCase()} · Fana Karaoke` : 'Fana Karaoke'}</title>
</svelte:head>

{#if ended}
  <main class="center-page">
    <div class="card end-card">
      <span class="badge badge-hot">{$tr('app.name')}</span>
      <h1>{$tr('tv.endedTitle')}</h1>
      <p class="muted">{$tr('tv.endedBody')}</p>
      <button class="btn btn-primary" type="button" on:click={newRoom}>{$tr('common.createRoom')}</button>
      <a class="btn btn-ghost" href="/">{$tr('common.home')}</a>
    </div>
  </main>
{:else if loadError}
  <main class="center-page">
    <div class="card end-card">
      <span class="badge">Fana</span>
      <h1>{errorText($tr, loadError)}</h1>
      <a class="btn btn-primary" href="/">{$tr('common.home')}</a>
    </div>
  </main>
{:else if !snap}
  <main class="center-page">
    <span class="spinner large"></span>
    <p class="muted">{$tr('common.loading')}</p>
  </main>
{:else}
  <div class="guest" class:host={isHost}>
    <header class="guest-header">
      <div class="topline">
        <a class="mini-brand" href="/">
          <span class="dot"></span>
          {$tr('app.name')}
        </a>
        <div class="row">
          <ConnectionDot {status} compact />
          <LanguageToggle />
          {#if !isHost}
            <button class="lock-btn" type="button" on:click={() => (pinSheet = true)} aria-label={$tr('room.hostControls')}>
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="5" y="11" width="14" height="9" rx="2"/>
                <path d="M8 11V8a4 4 0 0 1 8 0v3"/>
              </svg>
            </button>
          {:else}
            <span class="badge badge-cyan">{$tr('room.host')}</span>
          {/if}
        </div>
      </div>
      <div class="room-head">
        <div>
          <p class="label">{$tr('room.code')}</p>
          <h1 class="room-code">{snap.room.code.toUpperCase()}</h1>
        </div>
        <div class="room-stats">
          <span class="badge">{$tr('room.guests', { count: snap.room.guestCount, max: snap.room.guestCapacity })}</span>
          {#if snap.room.manualOrder}<span class="badge badge-cyan">{$tr('room.manualOrder')}</span>{/if}
        </div>
      </div>
      <div class="tabs" role="tablist">
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'search'}
          class:active={tab === 'search'}
          on:click={() => (tab = 'search')}
        >
          {$tr('room.tabSearch')}
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'queue'}
          class:active={tab === 'queue'}
          on:click={() => (tab = 'queue')}
        >
          {$tr('room.tabQueue')}
          <span class="count">{snap.queue.length}</span>
        </button>
      </div>
    </header>

    <main class="guest-main">
      {#if tab === 'search'}
        <section class="search-panel">
          <div class="search-box">
            <span class="search-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>
              </svg>
            </span>
            <input
              class="input search-input"
              bind:value={q}
              on:input={queueSearch}
              placeholder={$tr('room.searchPlaceholder')}
              aria-label={$tr('room.tabSearch')}
              autocomplete="off"
              enterkeyhint="search"
            />
            {#if q}
              <button class="clear-btn" type="button" on:click={clearSearch} aria-label={$tr('common.close')}>✕</button>
            {/if}
          </div>

          {#if locked}
            <div class="notice-strip">🔒 {$tr('room.queueLocked')}</div>
          {/if}
          {#if stale}<div class="notice-strip cyan">{$tr('room.stale')}</div>{/if}
          {#if searchFailed}
            <div class="notice-strip danger">{$tr('room.searchError')}</div>
          {/if}

          {#if searching}
            <div class="searching"><span class="spinner"></span>{$tr('room.searching')}</div>
          {:else if searched && results.length === 0}
            <div class="empty-search">
              <span>♪</span>
              <p>{$tr('room.noResults')}</p>
            </div>
          {:else if results.length > 0}
            <ul class="results">
              {#each results as result (result.videoId)}
                <li class="result card">
                  <img src={result.thumbnailUrl} alt="" loading="lazy" width="160" height="90" />
                  <div class="result-info">
                    <p class="result-title clamp-2">{result.title}</p>
                    <p class="tiny dim truncate">{result.channel} · {formatDuration(result.durationSeconds)}</p>
                  </div>
                  <button
                    class="btn btn-primary btn-sm"
                    type="button"
                    on:click={() => chooseResult(result)}
                    disabled={locked || snap.queue.length >= snap.room.queueCapacity}
                  >
                    {$tr('room.add')}
                  </button>
                </li>
              {/each}
            </ul>
          {:else}
            <div class="search-intro">
              <div class="mic-orb" aria-hidden="true">🎤</div>
              <p>{$tr('room.searchHint')}</p>
              <p class="tiny dim">{$tr('room.searchMin')}</p>
            </div>
          {/if}

          <form class="paste-card card" on:submit|preventDefault={preparePaste}>
            <p class="label">{$tr('room.pasteTitle')}</p>
            <div class="row">
              <input
                class="input grow"
                bind:value={pasteUrl}
                placeholder={$tr('room.pastePlaceholder')}
                inputmode="url"
                autocomplete="off"
                disabled={locked}
              />
              <button class="btn btn-secondary" type="submit" disabled={locked || snap.queue.length >= snap.room.queueCapacity}>{$tr('room.add')}</button>
            </div>
            {#if pasteError}<p class="error-text">{pasteError}</p>{/if}
          </form>
        </section>
      {:else}
        <section class="queue-panel">
          {#if snap.nowPlaying}
            {@const np = snap.nowPlaying}
            <article class="now-card card">
              <div class="now-bg" style={`background-image:url(${np.thumbnailUrl})`}></div>
              <div class="now-content">
                <p class="label">{$tr('room.nowPlaying')}</p>
                <h2 class="clamp-2">{np.title}</h2>
                <p class="small muted">
                  {$tr('tv.nowSinging')}: <strong>{np.nickname}</strong>
                  {#if np.mine}<span class="badge badge-hot">{$tr('room.youAre')}</span>{/if}
                </p>
                <div class="progress" aria-label={$tr('room.progress')}>
                  <span style={`width:${np.durationSeconds ? Math.min(100, (elapsedNow / np.durationSeconds) * 100) : 0}%`}></span>
                </div>
                <div class="spread tiny dim">
                  <span>{formatDuration(elapsedNow)}</span>
                  <span>{formatDuration(np.durationSeconds)}</span>
                </div>

                <div class="skip-zone">
                  {#if snap.skipVote?.active}
                    <div class="vote-live">
                      <div class="spread">
                        <strong>{$tr('tv.skipVoteTitle')}</strong>
                        <span class="mono">{voteLeft}s</span>
                      </div>
                      <p class="small">
                        {$tr('room.needed', { yes: snap.skipVote.yes, needed: snap.skipVote.threshold })}
                      </p>
                      <div class="vote-bar">
                        <span style={`width:${Math.min(100, (snap.skipVote.yes / Math.max(1, snap.skipVote.threshold)) * 100)}%`}></span>
                      </div>
                      {#if snap.skipVote.canVote}
                        <button class="btn btn-primary btn-sm" type="button" on:click={castSkip}>{$tr('room.voteYes')}</button>
                      {:else if snap.skipVote.votedByMe}
                        <span class="badge badge-hot">{$tr('room.votedYes')}</span>
                      {:else}
                        <p class="tiny dim">{$tr('room.votersOnly')}</p>
                      {/if}
                      {#if isHost}
                        <button class="btn btn-ghost btn-sm" type="button" on:click={cancelSkip}>
                          {$tr('tv.cancelVote')}
                        </button>
                      {/if}
                    </div>
                  {:else if cooldownLeft > 0}
                    <button class="btn btn-secondary btn-sm" type="button" disabled>
                      {$tr('room.cooldown', { seconds: cooldownLeft })}
                    </button>
                  {:else if !snap.you.canSkipVote}
                    <button class="btn btn-secondary btn-sm" type="button" disabled>
                      {$tr('room.voteToSkip')}
                    </button>
                    <p class="tiny dim">{$tr('room.votersOnly')}</p>
                  {:else}
                    <button class="btn btn-secondary btn-sm" type="button" on:click={startSkip}>
                      {$tr('room.voteToSkip')}
                    </button>
                  {/if}
                </div>
              </div>
            </article>
          {:else}
            <div class="empty-queue card">
              <span class="big-note">♪</span>
              <h2>{$tr('room.emptyTitle')}</h2>
              <p class="muted">{$tr('room.emptyBody')}</p>
              <button class="btn btn-primary" type="button" on:click={() => (tab = 'search')}>
                {$tr('room.tabSearch')}
              </button>
            </div>
          {/if}

          <div class="queue-heading">
            <h2>{$tr('room.upNext')}</h2>
            <span class="tiny dim">{$tr('room.queueCount', { count: snap.queue.length })}</span>
          </div>

          {#if queue.length === 0}
            <p class="tiny dim empty-line">{$tr('room.emptyBody')}</p>
          {/if}

          <ol class="queue-list">
            {#each queue as item, index (item.id)}
              <li class="queue-item card" class:mine={item.mine}>
                <span class="position mono">{index + 1}</span>
                <img src={item.thumbnailUrl} alt="" loading="lazy" width="96" height="54" />
                <div class="queue-info">
                  <p class="queue-title clamp-2">{item.title}</p>
                  <p class="tiny dim truncate">
                    {item.nickname}{#if item.mine} · <span class="you">{$tr('room.youAre')}</span>{/if}
                    · {formatDuration(item.durationSeconds)}
                  </p>
                </div>
                <div class="queue-actions">
                  <button
                    class:active={item.votedByMe}
                    class="vote-btn"
                    type="button"
                    on:click={() => toggleVote(item.id)}
                    aria-label={$tr(item.votedByMe ? 'room.voted' : 'room.vote')}
                    title={$tr(item.votedByMe ? 'room.voted' : 'room.vote')}
                  >
                    <svg viewBox="0 0 24 24" width="15" height="15" fill={item.votedByMe ? 'currentColor' : 'none'} stroke="currentColor" stroke-width="2">
                      <path d="M12 19V5M5 12l7-7 7 7"/>
                    </svg>
                    <span>{item.score}</span>
                  </button>
                  {#if isHost}
                    <div class="host-item-actions">
                      <button type="button" class="mini-action" on:click={() => moveSong(index, -1)} disabled={index === 0} aria-label={$tr('room.moveUp')}>↑</button>
                      <button type="button" class="mini-action" on:click={() => moveSong(index, 1)} disabled={index === queue.length - 1} aria-label={$tr('room.moveDown')}>↓</button>
                      <button type="button" class="mini-action danger" on:click={() => removeSong(item.id)} aria-label={$tr('room.remove')}>✕</button>
                    </div>
                  {/if}
                </div>
              </li>
            {/each}
          </ol>
        </section>
      {/if}
    </main>

    {#if isHost}
      <div class="host-bar">
        <div class="host-playback">
          {#if snap.nowPlaying?.status === 'playing'}
            <button class="hbtn" type="button" on:click={() => playback('pause')} disabled={busy} aria-label={$tr('tv.pause')}>
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><rect x="6" y="5" width="4" height="14" rx="1"/><rect x="14" y="5" width="4" height="14" rx="1"/></svg>
            </button>
          {:else}
            <button class="hbtn" type="button" on:click={() => playback('play')} disabled={busy} aria-label={$tr('tv.play')}>
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M7 5v14l12-7z"/></svg>
            </button>
          {/if}
          <button class="hbtn" type="button" on:click={() => playback('next')} disabled={busy} aria-label={$tr('tv.next')}>
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M6 5v14l9-7z"/><rect x="16" y="5" width="2" height="14" rx="1"/></svg>
          </button>
          {#if snap.nowPlaying?.status === 'countdown'}
            <button class="hbtn" type="button" on:click={() => playback('cancelCountdown')} disabled={busy} aria-label={$tr('tv.cancel')} title={$tr('tv.cancel')}>
              ✕
            </button>
          {/if}
        </div>
        <span class="host-mode-label">{$tr('room.hostMode')}</span>
        <div class="host-playback">
          <button class="hbtn" type="button" on:click={toggleLock} aria-label={$tr(locked ? 'room.unlock' : 'room.lock')}>
            <svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="5" y="11" width="14" height="9" rx="2"/>
              {#if locked}<path d="M8 11V8a4 4 0 0 1 7.5-1.9"/>{:else}<path d="M8 11V8a4 4 0 0 1 8 0v3"/>{/if}
            </svg>
          </button>
          <button class="btn btn-secondary btn-sm" type="button" on:click={() => (manageOpen = true)}>{$tr('room.manage')}</button>
        </div>
      </div>
    {/if}

    {#if toast}
      <div class="toast" role="status" aria-live="polite">{toast}</div>
    {/if}

    <Modal open={!!selected} title={$tr('room.nicknameTitle')} on:close={() => (selected = null)}>
      <div class="add-sheet">
        {#if selected?.title}
          <div class="selected-song">
            {#if selected.thumbnailUrl}<img src={selected.thumbnailUrl} alt="" />{/if}
            <div>
              <p class="clamp-2">{selected.title}</p>
              {#if selected.channel}<p class="tiny dim">{selected.channel}</p>{/if}
            </div>
          </div>
        {/if}
        <form on:submit|preventDefault={submitAdd} class="stack">
          <div class="field">
            <label for="nickname">{$tr('room.nicknameLabel')}</label>
            <input
              id="nickname"
              class="input"
              bind:value={nickname}
              placeholder={$tr('room.nicknamePlaceholder')}
              maxlength="24"
              autocomplete="off"
              use:focusOnMount
            />
            {#if nicknameError}<p class="error-text">{nicknameError}</p>{/if}
          </div>
          {#if addError}<p class="error-text" role="alert">{addError}</p>{/if}
          <div class="row end">
            <button class="btn btn-ghost" type="button" on:click={() => (selected = null)}>{$tr('common.cancel')}</button>
            <button class="btn btn-primary" type="submit" disabled={adding}>
              {#if adding}<span class="spinner"></span>{$tr('room.adding')}{:else}{$tr('room.submit')}{/if}
            </button>
          </div>
        </form>
      </div>
    </Modal>

    <Modal open={pinSheet} title={$tr('room.hostUnlockTitle')} on:close={() => (pinSheet = false)}>
      <form class="stack" on:submit|preventDefault={unlock}>
        <p class="muted">{$tr('room.hostUnlockBody')}</p>
        <div class="field">
          <label for="pin">{$tr('room.pin')}</label>
          <input
            id="pin"
            class="input input-mono pin-input"
            bind:value={pin}
            placeholder={$tr('room.pinPlaceholder')}
            maxlength="4"
            inputmode="numeric"
            pattern="[0-9]*"
            autocomplete="one-time-code"
            use:focusOnMount
          />
        </div>
        {#if pinError}<p class="error-text" role="alert">{pinError}</p>{/if}
        <button class="btn btn-primary" type="submit" disabled={unlocking || pin.trim().length !== 4}>
          {#if unlocking}<span class="spinner"></span>{/if}{$tr('room.unlockButton')}
        </button>
      </form>
    </Modal>

    <Modal open={manageOpen} title={$tr('room.manage')} on:close={() => (manageOpen = false)}>
      <div class="stack">
        <button class="btn btn-secondary" type="button" on:click={toggleLock}>
          {$tr(locked ? 'room.unlock' : 'room.lock')}
        </button>
        {#if !confirmClear}
          <button class="btn btn-secondary" type="button" on:click={() => (confirmClear = true)}>{$tr('room.clear')}</button>
        {:else}
          <div class="confirm-box">
            <p class="small">{$tr('room.clearConfirm')}</p>
            <div class="row end">
              <button class="btn btn-ghost btn-sm" type="button" on:click={() => (confirmClear = false)}>{$tr('common.cancel')}</button>
              <button class="btn btn-danger btn-sm" type="button" on:click={clearQueue}>{$tr('common.confirm')}</button>
            </div>
          </div>
        {/if}
        <hr />
        {#if !confirmEnd}
          <button class="btn btn-danger" type="button" on:click={() => (confirmEnd = true)}>{$tr('room.endRoom')}</button>
        {:else}
          <div class="confirm-box danger">
            <p class="small">{$tr('room.endConfirm')}</p>
            <div class="row end">
              <button class="btn btn-ghost btn-sm" type="button" on:click={() => (confirmEnd = false)}>{$tr('common.cancel')}</button>
              <button class="btn btn-danger btn-sm" type="button" on:click={endRoom}>{$tr('common.confirm')}</button>
            </div>
          </div>
        {/if}
      </div>
    </Modal>
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

  .center-page {
    min-height: 100dvh;
    display: grid;
    place-content: center;
    justify-items: center;
    gap: 1rem;
    padding: 2rem;
    text-align: center;
  }

  .end-card {
    padding: 2rem;
    display: grid;
    gap: 0.9rem;
    justify-items: center;
    max-width: 440px;
  }

  .end-card h1 {
    font-size: 1.8rem;
  }

  .end-card .btn {
    width: 100%;
  }

  .spinner.large {
    width: 30px;
    height: 30px;
    border-width: 3px;
    color: var(--magenta);
  }

  .guest-header {
    position: sticky;
    top: 0;
    z-index: 30;
    padding: 0.75rem 1rem 0;
    background: linear-gradient(to bottom, rgb(8 5 15 / 0.98) 70%, rgb(8 5 15 / 0.86));
    backdrop-filter: blur(12px);
    border-bottom: 1px solid var(--border);
  }

  .topline {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.8rem;
    margin-bottom: 0.8rem;
  }

  .mini-brand {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    color: var(--text);
    text-decoration: none;
    font-weight: 800;
    font-size: 0.86rem;
    letter-spacing: -0.02em;
  }

  .mini-brand .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--magenta);
    box-shadow: 0 0 12px var(--magenta);
  }

  .lock-btn {
    display: grid;
    place-items: center;
    width: 34px;
    height: 34px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted);
  }

  .lock-btn:hover {
    color: var(--text);
    border-color: var(--cyan);
  }

  .room-head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 0.8rem;
  }

  .room-code {
    font-family: var(--mono);
    font-size: clamp(1.6rem, 8vw, 2.2rem);
    letter-spacing: 0.08em;
    line-height: 1.1;
    color: var(--text);
  }

  .room-stats {
    display: flex;
    gap: 0.35rem;
    flex-wrap: wrap;
    justify-content: flex-end;
    padding-bottom: 0.2rem;
  }

  .tabs {
    display: flex;
    gap: 0.4rem;
  }

  .tabs button {
    flex: 1;
    position: relative;
    min-height: 44px;
    border: 0;
    border-bottom: 2px solid transparent;
    background: transparent;
    color: var(--dim);
    font-weight: 800;
    font-size: 0.9rem;
  }

  .tabs button.active {
    color: var(--text);
    border-color: var(--magenta);
  }

  .count {
    display: inline-grid;
    place-items: center;
    min-width: 20px;
    height: 20px;
    padding: 0 5px;
    border-radius: 999px;
    background: var(--surface-3);
    font-size: 0.7rem;
    margin-left: 4px;
  }

  .guest-main {
    padding: 1rem;
  }

  .search-box {
    position: relative;
    margin-bottom: 0.9rem;
  }

  .search-icon {
    position: absolute;
    left: 0.9rem;
    top: 50%;
    transform: translateY(-50%);
    color: var(--dim);
    display: grid;
  }

  .search-input {
    padding-left: 2.7rem;
    padding-right: 2.6rem;
    height: 52px;
    border-radius: 16px;
    font-size: 1rem;
  }

  .clear-btn {
    position: absolute;
    right: 0.65rem;
    top: 50%;
    transform: translateY(-50%);
    border: 0;
    background: var(--surface-3);
    color: var(--muted);
    border-radius: 50%;
    width: 28px;
    height: 28px;
  }

  .notice-strip {
    padding: 0.65rem 0.8rem;
    margin-bottom: 0.75rem;
    border-radius: 12px;
    background: rgb(255 45 149 / 0.1);
    border: 1px solid rgb(255 45 149 / 0.25);
    color: #ffc0dd;
    font-size: 0.82rem;
  }

  .notice-strip.cyan {
    background: rgb(34 211 238 / 0.08);
    border-color: rgb(34 211 238 / 0.25);
    color: #b8f6ff;
  }

  .notice-strip.danger {
    background: rgb(255 92 92 / 0.1);
    border-color: rgb(255 92 92 / 0.28);
    color: #ffb3b3;
  }

  .searching {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    color: var(--muted);
    padding: 1.2rem 0.2rem;
    font-size: 0.9rem;
  }

  .results {
    list-style: none;
    margin: 0 0 1rem;
    padding: 0;
    display: grid;
    gap: 0.55rem;
  }

  .result {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    padding: 0.55rem;
    border-radius: 14px;
  }

  .result img {
    width: 104px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 9px;
    flex: none;
    background: var(--surface-3);
  }

  .result-info {
    flex: 1;
    min-width: 0;
  }

  .result-title {
    font-size: 0.9rem;
    font-weight: 700;
    line-height: 1.3;
    margin-bottom: 0.2rem;
  }

  .result .btn {
    flex: none;
  }

  .empty-search,
  .empty-queue {
    display: grid;
    justify-items: center;
    gap: 0.65rem;
    text-align: center;
    padding: 2.2rem 1rem;
    color: var(--muted);
  }

  .empty-search span,
  .big-note {
    font-size: 2.2rem;
    color: var(--magenta);
  }

  .empty-search p {
    max-width: 34ch;
    line-height: 1.5;
    font-size: 0.9rem;
  }

  .search-intro {
    display: grid;
    justify-items: center;
    gap: 0.6rem;
    text-align: center;
    padding: 2.4rem 1rem;
    color: var(--muted);
  }

  .search-intro p {
    max-width: 36ch;
    line-height: 1.5;
    font-size: 0.92rem;
  }

  .mic-orb {
    display: grid;
    place-items: center;
    width: 72px;
    height: 72px;
    border-radius: 50%;
    background: radial-gradient(circle at 30% 30%, rgb(255 45 149 / 0.35), rgb(139 92 246 / 0.12));
    border: 1px solid var(--border-bright);
    box-shadow: 0 0 40px rgb(255 45 149 / 0.18);
    font-size: 1.8rem;
    margin-bottom: 0.3rem;
  }

  .paste-card {
    margin-top: 1.4rem;
    padding: 1rem;
    display: grid;
    gap: 0.55rem;
  }

  .paste-card .input {
    min-height: 44px;
    font-size: 0.9rem;
  }

  .queue-panel {
    display: grid;
    gap: 1rem;
  }

  .now-card {
    position: relative;
    overflow: hidden;
    padding: 1rem;
  }

  .now-bg {
    position: absolute;
    inset: 0;
    background-size: cover;
    background-position: center;
    opacity: 0.28;
    filter: blur(2px) saturate(1.2);
    transform: scale(1.05);
  }

  .now-card::after {
    content: "";
    position: absolute;
    inset: 0;
    background: linear-gradient(180deg, rgb(18 10 31 / 0.4), rgb(18 10 31 / 0.94));
  }

  .now-content {
    position: relative;
    z-index: 1;
    display: grid;
    gap: 0.55rem;
  }

  .now-content h2 {
    font-size: 1.25rem;
    line-height: 1.25;
  }

  .progress {
    height: 5px;
    border-radius: 999px;
    background: rgb(255 255 255 / 0.12);
    overflow: hidden;
    margin-top: 0.3rem;
  }

  .progress span {
    display: block;
    height: 100%;
    border-radius: 999px;
    background: linear-gradient(90deg, var(--magenta), var(--cyan));
  }

  .skip-zone {
    margin-top: 0.4rem;
  }

  .vote-live {
    display: grid;
    gap: 0.4rem;
    padding: 0.7rem;
    border-radius: 12px;
    background: rgb(255 45 149 / 0.08);
    border: 1px solid rgb(255 45 149 / 0.22);
  }

  .vote-bar {
    height: 6px;
    border-radius: 999px;
    background: rgb(255 255 255 / 0.12);
    overflow: hidden;
  }

  .vote-bar span {
    display: block;
    height: 100%;
    background: var(--magenta);
  }

  .queue-heading {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 1rem;
    margin-top: 0.4rem;
  }

  .queue-heading h2 {
    font-size: 1.05rem;
  }

  .empty-line {
    padding: 0.4rem 0;
  }

  .queue-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.55rem;
  }

  .queue-item {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    padding: 0.55rem;
    border-radius: 14px;
  }

  .queue-item.mine {
    border-color: rgb(255 45 149 / 0.35);
  }

  .position {
    width: 18px;
    flex: none;
    color: var(--dim);
    font-size: 0.72rem;
    text-align: center;
  }

  .queue-item img {
    width: 82px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 8px;
    flex: none;
    background: var(--surface-3);
  }

  .queue-info {
    flex: 1;
    min-width: 0;
  }

  .queue-title {
    font-size: 0.88rem;
    font-weight: 700;
    line-height: 1.3;
    margin-bottom: 0.2rem;
  }

  .you {
    color: var(--magenta);
    font-weight: 700;
  }

  .queue-actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.3rem;
    flex: none;
  }

  .vote-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    min-width: 46px;
    min-height: 34px;
    padding: 0 0.5rem;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted);
    font-size: 0.78rem;
    font-weight: 800;
  }

  .vote-btn.active {
    border-color: var(--magenta);
    background: rgb(255 45 149 / 0.14);
    color: #ffb3d7;
  }

  .host-item-actions {
    display: flex;
    gap: 2px;
  }

  .mini-action {
    width: 26px;
    height: 26px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted);
    font-size: 0.75rem;
    padding: 0;
  }

  .mini-action.danger {
    color: var(--danger);
  }

  .mini-action:disabled {
    opacity: 0.3;
  }

  .host-bar {
    position: fixed;
    left: 50%;
    bottom: 0;
    transform: translateX(-50%);
    width: min(720px, 100%);
    z-index: 40;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.6rem;
    padding: 0.6rem 1rem calc(0.6rem + env(safe-area-inset-bottom));
    background: rgb(18 10 31 / 0.97);
    border-top: 1px solid var(--border-bright);
    backdrop-filter: blur(14px);
  }

  .host-playback {
    display: flex;
    gap: 0.35rem;
  }

  .hbtn {
    display: grid;
    place-items: center;
    width: 42px;
    height: 42px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: var(--surface-2);
    color: var(--text);
  }

  .hbtn:hover {
    border-color: var(--magenta);
  }

  .host-mode-label {
    font-size: 0.66rem;
    font-weight: 800;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--magenta);
  }

  .pin-input {
    font-size: 1.8rem;
    text-align: center;
    letter-spacing: 0.5em;
    height: 62px;
  }

  .add-sheet {
    display: grid;
    gap: 1rem;
  }

  .selected-song {
    display: flex;
    gap: 0.7rem;
    align-items: center;
    padding: 0.6rem;
    border-radius: 12px;
    background: var(--surface);
    border: 1px solid var(--border);
  }

  .selected-song img {
    width: 82px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 7px;
  }

  .selected-song p {
    font-size: 0.88rem;
    font-weight: 700;
    line-height: 1.3;
  }

  .row.end {
    justify-content: flex-end;
  }

  .confirm-box {
    padding: 0.8rem;
    border-radius: 12px;
    border: 1px solid var(--border);
    background: var(--surface);
    display: grid;
    gap: 0.6rem;
  }

  .confirm-box.danger {
    border-color: rgb(255 92 92 / 0.4);
  }

  hr {
    border: 0;
    border-top: 1px solid var(--border);
    margin: 0.2rem 0;
  }

  @media (min-width: 680px) {
    .guest-header,
    .guest-main {
      padding-left: 1.5rem;
      padding-right: 1.5rem;
    }
    .result img {
      width: 148px;
    }
  }
</style>
