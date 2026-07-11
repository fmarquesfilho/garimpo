/**
 * PollingTimer — smart polling com visibility management e backoff.
 * Usa runes do Svelte 5 para state reativo.
 *
 * Usage:
 *   const poll = createPollingTimer({
 *     interval: 30000,
 *     onTick: async () => { ... }
 *   });
 *   poll.start();
 *   // poll.countdown, poll.paused, poll.status são reativos
 */

const DEFAULT_INTERVAL = 30000;
const MIN_INTERVAL = 10000;
const MAX_INTERVAL = 120000;
const BACKOFF_THRESHOLD = 3;
const STALE_THRESHOLD_MS = 5 * 60 * 1000; // 5 min

/**
 * @param {Object} opts
 * @param {number} [opts.interval] - Polling interval in ms (default 30000)
 * @param {() => Promise<void>} opts.onTick - Callback fired on each poll cycle
 * @param {() => Promise<void>} [opts.onFullRefresh] - Called when tab returns after >5min
 */
export function createPollingTimer({ interval, onTick, onFullRefresh }) {
	// Resolve interval: URL param > env > default
	let configuredInterval = resolveInterval(interval);

	let countdown = $state(Math.ceil(configuredInterval / 1000));
	let paused = $state(false);
	let status = $state('idle'); // idle | live | paused | offline
	let lastTickAt = $state(0);
	let consecutiveErrors = $state(0);

	let currentInterval = configuredInterval;
	let tickTimer = null;
	let countdownTimer = null;
	let hiddenAt = 0;

	function start() {
		if (tickTimer) return;
		status = 'live';
		countdown = Math.ceil(currentInterval / 1000);
		scheduleNext();
		startCountdown();
		listenVisibility();
	}

	function stop() {
		clearTimeout(tickTimer);
		clearInterval(countdownTimer);
		tickTimer = null;
		countdownTimer = null;
		status = 'idle';
	}

	function scheduleNext() {
		tickTimer = setTimeout(async () => {
			try {
				await onTick();
				lastTickAt = Date.now();
				consecutiveErrors = 0;
				if (currentInterval !== configuredInterval) {
					currentInterval = configuredInterval; // reset backoff
				}
				status = 'live';
			} catch {
				consecutiveErrors++;
				if (consecutiveErrors >= BACKOFF_THRESHOLD) {
					currentInterval = Math.min(currentInterval * 2, MAX_INTERVAL);
				}
				status = 'offline';
			}
			countdown = Math.ceil(currentInterval / 1000);
			if (!paused) scheduleNext();
		}, currentInterval);
	}

	function startCountdown() {
		countdownTimer = setInterval(() => {
			if (!paused && countdown > 0) {
				countdown--;
			}
		}, 1000);
	}

	function onVisibility() {
		if (document.visibilityState === 'hidden') {
			hiddenAt = Date.now();
			paused = true;
			status = 'paused';
			clearTimeout(tickTimer);
			tickTimer = null;
		} else {
			paused = false;
			const elapsed = Date.now() - hiddenAt;

			if (elapsed > STALE_THRESHOLD_MS && onFullRefresh) {
				onFullRefresh();
			} else {
				// Immediate tick on return
				onTick()
					.then(() => {
						consecutiveErrors = 0;
						status = 'live';
					})
					.catch(() => {
						status = 'offline';
					});
			}

			countdown = Math.ceil(currentInterval / 1000);
			scheduleNext();
		}
	}

	function listenVisibility() {
		if (typeof document !== 'undefined') {
			document.addEventListener('visibilitychange', onVisibility);
		}
	}

	function destroy() {
		stop();
		if (typeof document !== 'undefined') {
			document.removeEventListener('visibilitychange', onVisibility);
		}
	}

	return {
		get countdown() {
			return countdown;
		},
		get paused() {
			return paused;
		},
		get status() {
			return status;
		},
		get lastTickAt() {
			return lastTickAt;
		},
		start,
		stop,
		destroy
	};
}

function resolveInterval(explicit) {
	// URL param override (for debugging)
	if (typeof window !== 'undefined') {
		const params = new URLSearchParams(window.location.search);
		const fromUrl = params.get('intervalo');
		if (fromUrl) return clamp(Number(fromUrl));
	}

	// Env var
	const fromEnv = typeof import.meta !== 'undefined' ? import.meta.env?.VITE_POLL_INTERVAL_MS : undefined;
	if (fromEnv) return clamp(Number(fromEnv));

	// Explicit or default
	return clamp(explicit ?? DEFAULT_INTERVAL);
}

function clamp(v) {
	if (isNaN(v) || v < MIN_INTERVAL) return MIN_INTERVAL;
	if (v > MAX_INTERVAL) return MAX_INTERVAL;
	return v;
}
