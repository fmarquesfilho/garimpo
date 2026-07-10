/**
 * SvelteKit server-side instrumentation (experimental, since SvelteKit 2.31).
 *
 * This file is guaranteed to run BEFORE any application code is imported.
 * Configures OpenTelemetry with W3C trace context propagation.
 *
 * In Cloudflare Workers: propagation-only (no OTLP export — Workers don't have gRPC).
 * In Node dev (vite dev): can optionally export to local Jaeger if configured.
 *
 * SvelteKit will automatically create spans for:
 * - handle hooks (and sequences)
 * - Server load functions
 * - Form actions
 * - Remote functions
 */

import { BasicTracerProvider } from '@opentelemetry/sdk-trace-base';
import { Resource } from '@opentelemetry/resources';
import { W3CTraceContextPropagator } from '@opentelemetry/core';

const resource = new Resource({ 'service.name': 'garimpei-web-ssr' });
const provider = new BasicTracerProvider({ resource });

provider.register({
	propagator: new W3CTraceContextPropagator()
});
