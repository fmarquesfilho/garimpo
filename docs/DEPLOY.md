# Deploy do Garimpo

Como colocar no ar para sua esposa testar e dar feedback. O ponto que define a
arquitetura: o backend guarda os segredos da Shopee, então **precisa rodar num
servidor** — não dá para ser só site estático.

## Onde hospedar (a decisão)

| Opção | Custo | Esforço | Observação |
|-------|-------|---------|------------|
| **OCI Free Tier (VM ARM Ampere)** ✅ | grátis, sem dólar | médio (setup de VM) | recomendado: você já roda o Radio Casa 13 lá; tudo numa origem |
| Mac Mini de casa + Cloudflare Tunnel | grátis | baixo | depende da casa/internet ligadas; o Mini é estação de música |
| Fly.io / Render | cobra em dólar | baixo | free tiers limitados; serviços dormem |

**Recomendação: OCI Free Tier.** Uma VM Ampere A1 (até 4 OCPU / 24 GB, sempre
grátis) sobra para isto. O nginx serve o site estático **e** faz proxy de `/api`
para o Go — uma origem só, um TLS só, sem CORS.

```
navegador ──https──> nginx ──┬─ /        -> site estático (SvelteKit em /var/www/garimpo)
                             └─ /api/...  -> 127.0.0.1:8080 (garimpo-api, systemd)
                                                   └─ lê SHOPEE_* de /etc/garimpo/garimpo.env
```

## Passo a passo (uma vez)

1. **Criar a VM.** No console da OCI, crie uma instância **Ampere (ARM)**, imagem
   Ubuntu 22.04+. Na sub-rede, libere as portas 80 e 443 (ingress). No Ubuntu,
   `sudo ufw allow 80,443/tcp` se o firewall local estiver ativo.

2. **Apontar o domínio.** Crie um registro A (ex.: `garimpo.seudominio.com`) para
   o IP público da VM. (Sem domínio, dá para testar pelo IP, mas o TLS exige domínio.)

3. **Instalar o básico.**
   ```bash
   sudo apt update && sudo apt install -y nginx rsync
   ```

4. **Usuário e pastas.**
   ```bash
   sudo useradd --system --shell /usr/sbin/nologin garimpo
   sudo install -d -o garimpo -g garimpo /opt/garimpo
   sudo install -d -o <SEU_USER_DEPLOY> -g <SEU_USER_DEPLOY> /var/www/garimpo
   sudo install -d /etc/garimpo
   ```
   Deixe `/opt/garimpo` e `/var/www/garimpo` graváveis pelo usuário que o CI usa
   para o rsync (assim só o `systemctl` precisa de sudo).

5. **Segredos da Shopee.**
   ```bash
   sudo cp deploy/garimpo.env.example /etc/garimpo/garimpo.env
   sudo nano /etc/garimpo/garimpo.env      # preencha APP_ID e SECRET
   sudo chmod 600 /etc/garimpo/garimpo.env
   sudo chown garimpo:garimpo /etc/garimpo/garimpo.env
   ```

6. **systemd.**
   ```bash
   sudo cp deploy/garimpo-api.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable garimpo-api      # sobe sozinho no boot
   # (só inicia de fato depois do primeiro deploy colocar o binário)
   ```

7. **nginx.**
   ```bash
   sudo cp deploy/nginx-garimpo.conf /etc/nginx/sites-available/garimpo
   sudo sed -i 's/garimpo.SEUDOMINIO/garimpo.seudominio.com/' /etc/nginx/sites-available/garimpo
   sudo ln -s /etc/nginx/sites-available/garimpo /etc/nginx/sites-enabled/
   sudo nginx -t && sudo systemctl reload nginx
   ```

8. **TLS (Let's Encrypt).**
   ```bash
   sudo apt install -y certbot python3-certbot-nginx
   sudo certbot --nginx -d garimpo.seudominio.com
   ```

9. **Permitir o CI reiniciar os serviços.** Crie `/etc/sudoers.d/garimpo-deploy`
   (via `sudo visudo -f`) com a linha (troque `<SEU_USER_DEPLOY>`):
   ```
   <SEU_USER_DEPLOY> ALL=(root) NOPASSWD: /bin/systemctl restart garimpo-api, /bin/systemctl reload nginx, /usr/sbin/nginx -t
   ```

## Secrets do GitHub Actions

Em **Settings → Secrets and variables → Actions**, crie:

| Secret | O que é |
|--------|---------|
| `DEPLOY_HOST` | IP ou domínio da VM |
| `DEPLOY_USER` | usuário de deploy (o que tem o sudoers acima) |
| `DEPLOY_SSH_KEY` | **chave privada** SSH cuja pública está no `~/.ssh/authorized_keys` da VM |
| `DEPLOY_PORT` | opcional; só se o SSH não for a porta 22 |

Os segredos da Shopee **não** entram no GitHub — vivem só no `garimpo.env` da VM.

## Como o pipeline funciona

- **`ci.yml`** roda em todo push/PR: `go vet` + `go test` e build do front. É o
  portão — nada quebrado encosta na main.
- **`deploy.yml`** roda quando a main muda (ou manual): testa, compila o binário
  para a arquitetura da VM (`GOARCH=arm64` por padrão — mude para `amd64` se usar
  a VM AMD), gera o site estático e publica por SSH (rsync + restart do systemd +
  reload do nginx).

É o ciclo de DevOps da disciplina na prática: integração contínua dando feedback
rápido, e entrega contínua levando cada incremento testável ao ar. A primeira
ida ao ar é manual (os passos acima); daí em diante, `git push` na main publica.

## Primeiro deploy

Depois do setup, faça um push na main (ou rode o workflow **deploy** manualmente
pela aba Actions). Acompanhe o log; ao terminar, abra `https://garimpo.seudominio.com`.
Se a página subir mas a busca vier vazia, confira no servidor:
```bash
systemctl status garimpo-api
journalctl -u garimpo-api -n 50 --no-pager
curl -s localhost:8080/api/health
```

## Sobre o feedback dela

Como é a primeira pessoa real usando, vale deixar o caminho de retorno curto: um
canal simples (um grupo, um formulário, ou as próprias Issues do GitHub) onde ela
anota o que travou ou o que queria que existisse. Cada item vira um candidato a
incremento — que é exatamente como o resto do projeto vem sendo construído.
