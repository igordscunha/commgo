# Chat Offline P2P em Go

Uma aplicação de chat P2P que se descobre na rede local usando mDNS e troca mensagens via WebSockets.

---

## Fluxo de Execução

### 1. Início da Aplicação

- Ao subir, o servidor HTTP é iniciado na porta `0` — o sistema operacional escolhe uma porta livre.  
- Há um endpoint WebSocket disponível em `/ws` para receber conexões de outros pares.

---

### 2. Fase de Descoberta (mDNS)

- A aplicação anuncia o serviço `_meuchat._tcp` na porta alocada, incluindo metadados (por exemplo, nome de usuário).  
- Em paralelo, faz buscas pelo mesmo serviço na rede local para descobrir novos pares.

---

### 3. Gerenciamento de Pares

- Mantém um mapa em Go (`map[string]PeerInfo`) com informações de cada par descoberto.  
- Quando um par é encontrado:
  1. Adiciona ao mapa.  
  2. Abre uma conexão WebSocket de saída para `ws://<peer-endereço>/ws`.  
- Quando um par some (anúncio mDNS expira ou desconecta):
  1. Remove do mapa.  
  2. Fecha a conexão WebSocket correspondente.

---

### 4. Comunicação (WebSockets)

- **Envio de mensagens**  
  - Ao digitar um texto, a aplicação empacota em JSON e envia a **todas** as conexões de saída ativas.  
- **Recebimento de mensagens**  
  - O servidor WebSocket escuta em `/ws`.  
  - Mensagens recebidas de outros pares são exibidas na interface local.

---
