/**
 * OpenTelemetry trace propagation for the browser.
 *
 * Injects `traceparent` header into all fetch() calls to the backend.
 * This allows Cloud Trace to show traces that start in the browser.
 *
 * No spans are exported from the browser — only context propagation.
 * Zero cost, zero vendor dependency.
 *
 * Import this file as a side-effect in +layout.js to activate.
 */
import { browser } from '$app/environment';

if (browser) {
	initBrowserTracing();
}

async function initBrowserTracing() {
	try {
		const { WebTracerProvider } = await import('@opentelemetry/sdk-trace-web');
		const { W3CTraceContextPropagator } = await import('@opentelemetry/core');
		const { registerInstrumentations } = await import('@opentelemetry/instrumentation');
		const { FetchInstrumentation } = await import('@opentelemetry/instrumentation-fetch');

		const provider = new WebTracerProvider({
			resource: { attributes: { 'service.name': 'garimpei-web' } }
		});

		// Only propagation — no exporter (no spans sent from browser)
		provider.register({
			propagator: new W3CTraceContextPropagator()
		});

		registerInstrumentations({
			instrumentations: [
				new FetchInstrumentation({
					propagateTraceHeaderCorsUrls: [
						/garimpei\.app\.br/,
						/localhost/,
						/127\.0\.0\.1/
					],
					clearTimingResources: true
				})
			]
		});
	} catch {
		// OTel packages not installed or failed — graceful degradation
	}
}
