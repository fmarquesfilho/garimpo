/**
 * SvelteKit server-side instrumentation (experimental, since SvelteKit 2.31).
 *
 * Configures OpenTelemetry with W3C trace context propagation.
 * In Cloudflare Workers: propagation-only (no OTLP export).
 * In Node dev (vite dev): propagation + optional local collector export.
 *
 * SvelteKit auto-creates spans for: handle hooks, load functions, form actions.
 */

import { trace, propagation } from '@opentelemetry/api';
import { BasicTracerProvider } from '@opentelemetry/sdk-trace-base';
import { resourceFromAttributes } from '@opentelemetry/resources';
import { W3CTraceContextPropagator } from '@opentelemetry/core';

const resource = resourceFromAttributes({ 'service.name': 'garimpei-web-ssr' });
const provider = new BasicTracerProvider({ resource });

trace.setGlobalTracerProvider(provider);
propagation.setGlobalPropagator(new W3CTraceContextPropagator());
