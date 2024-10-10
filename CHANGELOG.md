# Changelog

## Versão 2.0.0 - Data: 09 de outubro de 2024

### Novas Funcionalidades

- **Suporte a Contexto nas Conversas com OpenAI:**
    - Atualizamos o cliente OpenAI para enviar o histórico completo da conversa, permitindo que o modelo mantenha o contexto entre as mensagens.
    - Modificamos o método `SendPrompt` em `OpenAIClient` para aceitar um novo parâmetro `history` e incluí-lo no payload enviado à API.
    - Ajustamos a interface `LLMClient` para refletir a nova assinatura do método `SendPrompt(prompt string, history []models.Message)`.

- **Seleção Dinâmica de Modelos OpenAI:**
    - Agora é possível selecionar o modelo da OpenAI a ser utilizado (`gpt-3.5-turbo`, `gpt-4`, etc.) através da variável de ambiente `OPENAI_MODEL`.
    - O nome do assistente é exibido dinamicamente no frontend com base no modelo selecionado.

- **Atualizações no Frontend:**
    - Alteramos o `script.js` para armazenar e enviar o histórico completo da conversa ao servidor, garantindo que o contexto seja mantido.
    - Criamos a função `getConversationHistory()` para extrair e formatar o histórico das mensagens armazenadas no `localStorage`.
    - Ajustamos a função `sendMessageToServer()` para incluir o histórico na requisição ao backend.
    - Atualizamos o `index.html` para expor o `modelName` ao JavaScript através do atributo `data-model-name`.

### Melhorias

- **Compatibilidade com StackSpot AI:**
    - Atualizamos o `StackSpotClient` para implementar a nova interface `LLMClient`, aceitando o parâmetro `history` no método `SendPrompt`, garantindo compatibilidade.
    - Implementamos lógica para ignorar o histórico ou concatená-lo ao prompt, permitindo que o aplicativo funcione corretamente com ambos os provedores.

- **Correções de Bugs:**
    - **Incompatibilidade de Tipos:**
        - Resolvemos o erro de incompatibilidade de tipos ao mover a definição de `Message` para um pacote comum `models`, garantindo que `handlers` e `llm` utilizem o mesmo tipo.
    - **Parâmetro `model` Ausente:**
        - Adicionamos o campo `model` ao `OpenAIClient` e incluímos o parâmetro `model` no payload enviado à API da OpenAI, corrigindo o erro onde o parâmetro `model` não estava sendo enviado.
    - **Variável `modelName` Não Utilizada:**
        - Corrigimos o erro onde a variável `modelName` foi declarada mas não estava sendo utilizada no `main.go`, assegurando que seja passada ao `indexHandler` e utilizada no template.

- **Interface do Usuário:**
    - Ajustamos o `script.js` e o `index.html` para exibir dinamicamente o nome do assistente com base no provedor e no modelo selecionados.
    - Melhoramos a experiência do usuário ao exibir o nome correto do modelo (por exemplo, `GPT-4` ou `ChatGPT`) durante as conversas.

- **Tratamento de Erros e Logs:**
    - Melhoramos o tratamento de erros no backend, fornecendo mensagens mais claras em caso de falhas na comunicação com os provedores de LLM.
    - Adicionamos logs detalhados para facilitar a depuração e monitoramento da aplicação.

### Documentação

- **Atualização do README.md:**
    - Expandimos o README para incluir instruções detalhadas sobre como configurar e usar o aplicativo com os provedores StackSpot AI e OpenAI.
    - Adicionamos seções sobre as novas funcionalidades, como manutenção de contexto e seleção dinâmica de provedores e modelos.
    - Incluímos detalhes técnicos sobre as mudanças na arquitetura e no fluxo de dados entre frontend e backend.
    - Atualizamos as capturas de tela para refletir as alterações na interface e nas funcionalidades.
    - Fornecemos referências para a documentação oficial da StackSpot AI e da OpenAI.

### Outras Alterações

- **Código Limpo e Organizado:**
    - Refatoramos partes do código para melhorar a legibilidade e manutenção.
    - Comentamos funções e trechos de código para melhor compreensão.

- **Suporte a Múltiplos Provedores:**
    - A estrutura do código foi aprimorada para facilitar a adição de novos provedores de LLM no futuro.

---

## Versão 1.0.0 - Data: 08 de outubro de 2024

- **Lançamento inicial do aplicativo de chat com integração à StackSpot AI.**
    - Funcionalidades básicas de chat interativo com suporte a múltiplas conversas.
    - Integração com StackSpot AI utilizando Quick Commands e Agentes Especializados.
    - Interface amigável com suporte a Markdown e realce de sintaxe para código.
    - Armazenamento de histórico de conversas no `localStorage`.

---

**Notas Importantes:**

- **Configuração Necessária:**
    - Para utilizar as novas funcionalidades, especialmente a manutenção de contexto com a OpenAI, é necessário atualizar as variáveis de ambiente e certificar-se de que sua chave de API tem acesso ao modelo especificado.
    - Verifique o arquivo `README.md` para obter instruções detalhadas sobre a configuração.

- **Compatibilidade:**
    - As alterações mantêm a compatibilidade com a StackSpot AI, garantindo que os usuários possam escolher o provedor de LLM que melhor atenda às suas necessidades.

- **Monitoramento de Custos:**
    - O uso de modelos como o `gpt-4` pode implicar em custos maiores. Recomenda-se monitorar o uso da API da OpenAI para evitar surpresas na fatura.

---

Para quaisquer dúvidas ou problemas, por favor, consulte a seção de [Resolução de Problemas](README.md#resolução-de-problemas) no README ou abra uma issue no repositório.