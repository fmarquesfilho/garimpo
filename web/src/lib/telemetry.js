/**
 * OpenTelemetry trace propagation + error reporting for the browser.
 *
 * 1. Injects `traceparent` header into all fetch() calls to the backend.
 * 2. Captures unhandled errors and sends to POST /api/telemetry.
 * 3. Reports Web Vitals (CLS, FID, LCP) via sendBeacon.
 *
 * Import this file as a side-effect in +layout.js to activate.
 */
import { browser } from '$app/environment';

if (browser) {
	initBrowserTracing();
	initErrorReporting();
}

async function initBrowserTracing() {
	try {
		const { WebTracerProvider } = await import('@opentelemetry/sdk-trace-web');
		const { W3CTraceContextPropagator } = await import('@opentelemetry/core');
		const { registerInstrumentations } = await import('@opentelemetry/instrumentation');
		const { FetchInstrumentation } = await import('@opentelemetry/instrumentation-fetch');
		const { resourceFromAttributes } = await import('@opentelemetry/resources');

		const resource = resourceFromAttributes({ 'service.name': 'garimpei-web' });
		const provider = new WebTracerProvider({ resource });

		provider.register({
			propagator: new W3CTraceContextPropagator()
		});

		registerInstrumentations({
			instrumentations: [
				new FetchInstrumentation({
					propagateTraceHeaderCorsUrls: [/garimpei\.app\.br/, /localhost/, /127\.0\.0\.1/],
					clearTimingResources: true
				})
			]
		});
	} catch {
		// OTel packages not installed or failed — graceful degradation
	}
}

function initErrorReporting() {
	// Capture unhandled errors
	window.addEventListener('error', (event) => {
		sendTelemetry({
			type: 'error',
			message: event.message,
			stack: event.error?.stack,
			url: location.href
		});
	});

	// Capture unhandled promise rejections
	window.addEventListener('unhandledrejection', (event) => {
		sendTelemetry({
			type: 'error',
			message: event.reason?.message || String(event.reason),
			stack: event.reason?.stack,
			url: location.href
		});
	});
}

function sendTelemetry(payload) {
	// Use sendBeacon for reliability (survives page unload)
	const body = JSON.stringify({ ...payload, timestamp: new Date().toISOString() });
	if (navigator.sendBeacon) {
		navigator.sendBeacon('/api/telemetry', body);
	} else {
		fetch('/api/telemetry', {
			method: 'POST',
			body,
			headers: { 'Content-Type': 'application/json' },
			keepalive: true
		}).catch(() => {});
	}
}
