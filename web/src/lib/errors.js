/**
 * Tratamento de erros estruturados do Garimpei (frontend).
 *
 * O backend retorna Problem Details (RFC 9457) com:
 *   { type, title, status, detail, code, retry }
 *
 * Este módulo classifica erros por tipo/status e fornece mensagens
 * amigáveis para cada classe de falha.
 */

// ── Classificação de erros ───────────────────────────────────────────────────

/**
 * Retorna true se o erro é de autenticação (401).
 * O app deve redirecionar para login.
 */
export function isAuthError(err) {
	return err?.status === 401;
}

/**
 * Retorna true se o erro é de permissão (403).
 */
export function isForbiddenError(err) {
	return err?.status === 403;
}

/**
 * Retorna true se o recurso não foi encontrado (404).
 */
export function isNotFoundError(err) {
	return err?.status === 404;
}

/**
 * Retorna true se o erro é de validação/input (400).
 */
export function isValidationError(err) {
	return err?.status === 400;
}

/**
 * Retorna true se o erro indica falha em serviço externo (502/503).
 */
export function isExternalServiceError(err) {
	return err?.status === 502 || err?.status === 503;
}

/**
 * Retorna true se o erro pode ser retentado (backend indica retry: true).
 */
export function isRetryable(err) {
	return err?.retry === true || err?.problem?.retry === true;
}

/**
 * Retorna true se é um erro de rede (fetch falhou antes de receber resposta).
 */
export function isNetworkError(err) {
	return err?.name === 'TypeError' || err?.name === 'AbortError';
}

// ── Mensagens amigáveis ──────────────────────────────────────────────────────

const MENSAGENS_PADRAO = {
	401: 'Sessão expirada. Faça login novamente.',
	403: 'Você não tem permissão para esta ação.',
	404: 'Recurso não encontrado.',
	409: 'Conflito — verifique se o item já existe.',
	500: 'Erro interno. Tente novamente em instantes.',
	502: 'Serviço externo indisponível. Tente novamente.',
	503: 'Serviço temporariamente indisponível.'
};

/**
 * Extrai uma mensagem amigável do erro para exibir ao usuário.
 * Prioridade: detail do Problem → erro do Problem → mensagem padrão por status.
 */
export function mensagemAmigavel(err) {
	if (!err) return 'Erro desconhecido.';

	// Erros de rede
	if (isNetworkError(err)) {
		return 'Sem conexão com o servidor. Verifique sua internet.';
	}

	// Mensagem do Problem Details (já vem do backend)
	if (err.message && err.status) {
		return err.message;
	}

	// Fallback por status
	if (err.status && MENSAGENS_PADRAO[err.status]) {
		return MENSAGENS_PADRAO[err.status];
	}

	return err.message || 'Erro inesperado.';
}

/**
 * Código interno do erro (para lógica do frontend, não para exibição).
 * Ex.: "nao_autenticado", "servico_externo", "entrada_invalida"
 */
export function codigoErro(err) {
	return err?.code || err?.problem?.code || '';
}
