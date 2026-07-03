// Formatadores compartilhados usados em múltiplas páginas.
// Centraliza lógica de formatação para evitar duplicação.

/** Formata valor em BRL (ex.: R$ 89,90). */
export const brl = (v) => (v ?? 0).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });

/** Formata fração como porcentagem (ex.: 0.15 → "15%"). */
export const pct = (v) => `${((v ?? 0) * 100).toLocaleString('pt-BR', { maximumFractionDigits: 1 })}%`;

/** Formata número com separador de milhar, sem decimais. */
export const num = (v) => (v ?? 0).toLocaleString('pt-BR', { maximumFractionDigits: 0 });

/** Formata porcentagem com sinal (ex.: 0.05 → "+5.0%", -0.2 → "-20.0%"). */
export const pctSinal = (v) => {
	const val = ((v ?? 0) * 100).toFixed(1);
	return v >= 0 ? `+${val}%` : `${val}%`;
};

/** Formata ISO timestamp para dd/mm/yy HH:mm. */
export const dataHora = (v) => {
	if (!v) return '—';
	const d = new Date(v);
	if (isNaN(d.getTime())) return v;
	return d.toLocaleString('pt-BR', {
		day: '2-digit',
		month: '2-digit',
		year: '2-digit',
		hour: '2-digit',
		minute: '2-digit'
	});
};

/** Formata ISO timestamp para dd/mm/yyyy HH:mm. */
export const dataHoraCompleta = (v) => {
	if (!v) return '';
	const d = new Date(v);
	if (isNaN(d.getTime())) return v;
	return d.toLocaleString('pt-BR', {
		day: '2-digit',
		month: '2-digit',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit'
	});
};

/** Calcula tempo relativo (ex.: "5min atrás", "2h atrás"). */
export const tempoAtras = (v) => {
	if (!v) return '';
	const diff = Date.now() - new Date(v).getTime();
	const min = Math.floor(diff / 60000);
	if (min < 1) return 'agora';
	if (min < 60) return `${min}min atrás`;
	const h = Math.floor(min / 60);
	if (h < 24) return `${h}h atrás`;
	const d = Math.floor(h / 24);
	return `${d}d atrás`;
};

/** Extrai apenas a data (YYYY-MM-DD) de um ISO timestamp. */
export const apenasData = (v) => v?.split('T')[0] ?? '';
