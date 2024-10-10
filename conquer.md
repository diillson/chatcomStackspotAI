# Aplicativo de Chat com Integração à StackSpot AI e OpenAI

Bem-vindo ao aplicativo de chat interativo semelhante ao ChatGPT, mas integrado à **StackSpot AI** e aos modelos GPT da **OpenAI**. Este aplicativo permite que você se comunique com uma inteligência artificial em um ambiente web amigável, aproveitando os poderosos recursos fornecidos pela StackSpot AI e pela OpenAI, incluindo fontes de conhecimento, comandos rápidos, agentes especializados e conversas com manutenção de contexto.

## Sumário

- [Funcionalidades](#funcionalidades)
- [Pré-requisitos](#pré-requisitos)
- [Instalação e Configuração](#instalação-e-configuração)
- [Uso](#uso)
- [Integração com a StackSpot AI e OpenAI](#integração-com-a-stackspot-ai-e-openai)
    - [Provedores de LLM](#provedores-de-llm)
    - [Fontes de Conhecimento (StackSpot AI)](#fontes-de-conhecimento-stackspot-ai)
    - [Comandos Rápidos (StackSpot AI)](#comandos-rápidos-stackspot-ai)
    - [Agentes Especializados (StackSpot AI)](#agentes-especializados-stackspot-ai)
    - [Manutenção de Contexto (OpenAI)](#manutenção-de-contexto-openai)
- [Detalhes Técnicos](#detalhes-técnicos)
- [Resolução de Problemas](#resolução-de-problemas)
- [Contribuição](#contribuição)
- [Licença](#licença)
- [Agradecimentos](#agradecimentos)
- [Screenshots](#screenshots)
- [Referências](#referências)

## Funcionalidades

- **Chat Interativo:** Converse com uma inteligência artificial em tempo real, alimentada pela StackSpot AI ou pela OpenAI.
- **Múltiplas Conversas:** Crie, renomeie e exclua chats independentes.
- **Manutenção de Contexto:** O assistente de IA mantém o contexto da conversa para interações mais coerentes (ao usar OpenAI).
- **Troca Dinâmica de Provedor de LLM:** Alterne entre StackSpot AI e OpenAI em tempo de execução, sem reiniciar a aplicação.
- **Barra Lateral Personalizável:** Oculte ou exiba a barra lateral conforme sua preferência.
- **Histórico de Mensagens:** O histórico é armazenado no `localStorage` do navegador.
- **Suporte a Markdown:** Envie e receba mensagens formatadas em Markdown, com realce de sintaxe para código.
- **Indicador de Carregamento:** Enquanto a IA processa sua mensagem, um indicador "Pensando..." é exibido.
- **Interface Responsiva:** Design adaptável para diversos tamanhos de tela.
- **Segurança Integrada:** Sanitização de conteúdo para prevenir execução de código malicioso.
- **Seleção de Modelo (OpenAI):** Configure o aplicativo para usar diferentes modelos da OpenAI, como `gpt-3.5-turbo` ou `gpt-4`.

## Pré-requisitos

- **Go:** Versão 1.20+ instalada em sua máquina.
- **Navegador Moderno:** Google Chrome, Mozilla Firefox, Microsoft Edge ou equivalente.
- **Acesso a Provedores de LLM:**
    - **Para StackSpot AI:**
        - Conta com acesso às APIs da StackSpot AI.
        - `CLIENT_ID`, `CLIENT_SECRET` e `SLUG_NAME` para autenticação e acesso à API.
    - **Para OpenAI:**
        - Chave de API da OpenAI com acesso aos modelos desejados (por exemplo, `gpt-3.5-turbo`, `gpt-4`).
        - Observe que o acesso ao `gpt-4` pode exigir permissões adicionais.
- **Chaves de API:** Chaves de API e variáveis de ambiente configuradas corretamente para os provedores de LLM.

## Instalação e Configuração

### 1. Clone o Repositório

```bash
git clone https://github.com/diillson/chatcomStackspotAI.git
```

### 2. Navegue até o Diretório do Projeto

```bash
cd chatcomStackspotAI/
```

### 3. Configurar Variáveis de Ambiente

Defina as variáveis de ambiente para os provedores de LLM que deseja utilizar.

#### Para StackSpot AI:

- **CLIENT_ID:** Seu `client_id` da StackSpot AI.
- **CLIENT_SECRET:** Seu `client_secret` da StackSpot AI.
- **SLUG_NAME:** O slug do seu Quick Command ou agente.

Exemplo:

```bash
export CLIENT_ID=seu_client_id
export CLIENT_SECRET=seu_client_secret
export SLUG_NAME=seu_slug_name
```

#### Para OpenAI:

- **OPENAI_API_KEY:** Sua chave de API da OpenAI.
- **OPENAI_MODEL:** O modelo que você deseja usar (`gpt-3.5-turbo`, `gpt-4`, etc.).

Exemplo:

```bash
export OPENAI_API_KEY=sua_chave_api_openai
export OPENAI_MODEL=gpt-4  # ou gpt-3.5-turbo
```

**Nota:** Certifique-se de que suas chaves de API têm acesso aos modelos especificados.

### 4. Instale as Dependências Backend

```bash
go mod tidy
```

### 5. Execute o Servidor Backend

```bash
go run main.go
```

O servidor iniciará na porta `8080` por padrão.

### 6. Acesse o Aplicativo no Navegador

Abra o navegador e visite:

```
http://localhost:8080
```

## Uso

### Criar Nova Conversa

- Clique no botão **"Nova Conversa"** na barra lateral para iniciar um novo chat.
- A conversa será adicionada à lista na barra lateral.

### Enviar Mensagens

- Digite sua mensagem no campo de entrada na parte inferior.
- Pressione **"Enviar"** ou aperte **Enter** para enviar a mensagem.
- Aguarde a resposta da IA, que é fornecida pela StackSpot AI ou pela OpenAI, dependendo de sua configuração.
- O aplicativo mantém o contexto da conversa ao usar a OpenAI, permitindo interações mais coerentes.

### Alternar Entre Conversas

- Na barra lateral, clique no nome da conversa para alternar entre chats.
- Cada conversa mantém seu próprio histórico de mensagens.

### Renomear Conversas

- Clique no ícone de **lápis (✏️)** ao lado do nome da conversa na barra lateral.
- Insira o novo nome e confirme.

### Deletar Conversas

- Clique no ícone de **lixeira (🗑️)** ao lado do nome da conversa.
- Confirme a exclusão na janela que aparecerá.
- **Atenção:** Esta ação é irreversível e o histórico será perdido.

### Ocultar/Exibir Barra Lateral

- Use o botão de alternância **(⬅/➡)** no canto superior esquerdo para ocultar ou exibir a barra lateral.
- Isto é útil para maximizar a área de visualização do chat.

### Limpar Histórico

- Dentro de uma conversa, clique no botão **"Limpar Histórico"** para apagar todas as mensagens daquela conversa.

### Trocar o Provedor de LLM em Tempo de Execução

- No topo da página, você encontrará um menu suspenso que permite selecionar o provedor de LLM desejado.
- Selecione entre **StackSpotAI** e **OpenAI**.
- Ao alterar o provedor, a aplicação atualizará automaticamente para utilizar o novo provedor selecionado.
- **Observação:** Certifique-se de que as chaves de API e configurações para ambos os provedores estejam corretamente definidas, conforme explicado na seção [Instalação e Configuração](#instalação-e-configuração).

## Integração com a StackSpot AI e OpenAI

[... conteúdo permanece o mesmo ...]

## Detalhes Técnicos

### Modificações para Suporte à Troca Dinâmica de Provedor de LLM

- **LLMManager:** Implementação de uma estrutura que gerencia múltiplos clientes LLM e permite a troca dinâmica do provedor.
- **Endpoints Atualizados:**
    - **`/change-provider`:** Novo endpoint que recebe solicitações para alterar o provedor de LLM em tempo de execução.
- **Atualizações no Frontend:**
    - Adicionado um seletor (`select`) no `index.html` para permitir que o usuário escolha o provedor de LLM.
    - `script.js` atualizado para lidar com a mudança de provedor e recarregar a interface adequadamente.
- **Considerações sobre Concorrência:**
    - Uso de mutexes para garantir que a alteração do provedor seja thread-safe.
    - Garantia de que as instâncias dos clientes LLM são thread-safe.

## Resolução de Problemas

### Provedor de LLM Não Altera

- **Sintomas:** Ao selecionar um novo provedor de LLM, a aplicação continua utilizando o provedor anterior.
- **Soluções:**
    - Verifique se as variáveis de ambiente para ambos os provedores estão definidas e acessíveis pela aplicação.
    - Certifique-se de que o cliente para o provedor selecionado foi inicializado corretamente.
    - Confira se não há erros no console do navegador ou nos logs do servidor que possam indicar problemas na mudança de provedor.
    - Limpe o cache do navegador ou faça um recarregamento forçado da página.

[... demais conteúdos permanecem os mesmos ...]

---

## Agradecimentos

[Sem alterações]

---

# Changelog Atualizado

# Changelog

## Versão 1.2.0 - Data: 09 de outubro de 2024

### Novas Funcionalidades

- **Troca Dinâmica de Provedor de LLM:**
    - Implementamos a capacidade de alternar entre os provedores StackSpot AI e OpenAI em tempo de execução, sem a necessidade de reiniciar a aplicação.
    - Adicionamos a estrutura `LLMManager` para gerenciar múltiplos clientes LLM e permitir a seleção dinâmica do provedor.
    - Criamos o endpoint `/change-provider` para receber solicitações de alteração de provedor.
    - Atualizamos o frontend para incluir um menu suspenso que permite aos usuários selecionar o provedor de LLM desejado.

### Melhorias

- **Correções de Bugs:**
    - Corrigimos erros de digitação no nome do modelo OpenAI (`gpt-4o` para `gpt-4`), garantindo que o modelo correto seja utilizado.
    - Adicionamos logs adicionais nos handlers e no `LLMManager` para facilitar a depuração e monitoramento da aplicação.
    - Asseguramos que as variáveis de ambiente necessárias para ambos os provedores estejam definidas e acessíveis pela aplicação.

- **Interface do Usuário:**
    - Ajustamos o `script.js` e o `index.html` para refletir a seleção do provedor atual e atualizar dinamicamente o nome do assistente.

### Documentação

- **Atualização do README.md:**
    - Incluímos instruções detalhadas sobre como utilizar a nova funcionalidade de troca dinâmica de provedores.
    - Atualizamos as seções de funcionalidades e uso para refletir as mudanças implementadas.
    - Adicionamos informações técnicas sobre a implementação da troca dinâmica do provedor de LLM.

---

## Versão 1.1.0 - Data: 09 de outubro de 2024

[... conteúdo anterior ...]

---

## Versão 1.0.0 - Data: 08 de outubro de 2024

[... conteúdo anterior ...]

---

**Notas Importantes:**

- **Configuração Necessária:**
    - Para utilizar as novas funcionalidades, especialmente a troca dinâmica de provedores, é necessário atualizar as variáveis de ambiente e certificar-se de que suas chaves de API têm acesso aos modelos especificados.
    - Verifique o arquivo `README.md` para obter instruções detalhadas sobre a configuração.

- **Compatibilidade:**
    - As alterações mantêm a compatibilidade com ambos os provedores, garantindo que os usuários possam escolher o provedor de LLM que melhor atenda às suas necessidades.

- **Monitoramento de Custos:**
    - O uso de modelos como o `gpt-4` pode implicar em custos maiores. Recomenda-se monitorar o uso da API da OpenAI para evitar surpresas na fatura.

---

Para quaisquer dúvidas ou problemas, por favor, consulte a seção de [Resolução de Problemas](README.md#resolução-de-problemas) no README ou abra uma issue no repositório.

---