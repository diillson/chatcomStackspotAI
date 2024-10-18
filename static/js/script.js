// script.js

(function () {
    'use strict';

    // Verifica se o script foi carregado corretamente
    console.log("script.js carregado corretamente.");

    document.addEventListener('DOMContentLoaded', () => {
        // Variáveis do DOM
        const chatListDiv = document.getElementById('chat-list');
        const chatForm = document.getElementById('chat-form');
        const userInput = document.getElementById('user-input');
        const messagesDiv = document.getElementById('messages');
        const clearHistoryButton = document.getElementById('clear-history-button');
        const newChatButton = document.getElementById('new-chat-button');
        const toggleSidebarButton = document.getElementById('toggle-sidebar');
        const sidebar = document.getElementById('sidebar');
        const llmProviderSelect = document.getElementById('llm-provider-select');
        // Removemos estas linhas, pois não precisamos mais dos atributos data-llm-provider e data-model-name
        // const llmProvider = document.body.dataset.llmProvider;
        // const modelName = document.body.dataset.modelName;

        // Elemento para o Highlight.js
        const highlightStyleLink = document.getElementById('highlight-style');

        // Estado do aplicativo
        let currentChatID = null;
        // Carregamos o provedor do localStorage ou usamos 'STACKSPOT' como padrão
        let llmProvider = localStorage.getItem('llmProvider') || 'STACKSPOT';
        let modelName = ''; // O modelName pode ser definido se necessário
        let assistantName = getAssistantName(llmProvider, modelName);
        let shouldAutoScroll = true;

        // Função alternativa para gerar UUID
        function generateUUID() {
            let d = new Date().getTime();
            if (typeof performance !== 'undefined' && typeof performance.now === 'function') {
                d += performance.now(); // Use high-precision timer if available
            }
            return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
                const r = (d + Math.random() * 16) % 16 | 0;
                d = Math.floor(d / 16);
                return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
            });
        }

        // Inicialização
        initialize();

        // Função principal de inicialização
        function initialize() {
            // Configurar o seletor de provedor LLM
            llmProviderSelect.value = llmProvider;

            // Atualizar o assistantName
            assistantName = getAssistantName(llmProvider, modelName);

            // Inicializar Highlight.js
            if (typeof hljs !== 'undefined') {
                hljs.highlightAll();
            } else {
                console.error("hljs não está definido. Verifique a inclusão da biblioteca.");
            }

            // Carregar a lista de chats
            loadChatList();

            // Selecionar o chat atual ou criar um novo
            const storedCurrentChat = localStorage.getItem('currentChatID');
            if (storedCurrentChat && isChatExists(storedCurrentChat)) {
                currentChatID = storedCurrentChat;
                loadChatHistory();
            } else {
                currentChatID = createNewChat();
                loadChatHistory();
            }

            // Event listeners
            addEventListeners();
        }

        // Adiciona todos os event listeners necessários
        function addEventListeners() {
            llmProviderSelect.addEventListener('change', handleProviderChange);
            chatForm.addEventListener('submit', handleFormSubmit);
            userInput.addEventListener('keydown', handleUserInputKeyDown);
            userInput.addEventListener('input', debounce(autoResizeTextarea, 50));
            messagesDiv.addEventListener('scroll', throttle(handleMessagesScroll, 100));
            clearHistoryButton.addEventListener('click', clearChatHistory);
            newChatButton.addEventListener('click', handleNewChat);
            toggleSidebarButton.addEventListener('click', toggleSidebar);

            // Adiciona o event listener para o botão de alternância de tema
            const toggleThemeButton = document.getElementById('toggle-theme');
            toggleThemeButton.addEventListener('click', toggleTheme);

            // Carrega o tema preferido do usuário
            loadUserTheme();
        }

        // Obtém o nome do assistente com base no provedor e modelo
        function getAssistantName(provider, model) {
            switch (provider) {
                case 'STACKSPOT':
                    return 'StackSpotAI';
                case 'OPENAI':
                    switch (model) {
                        case 'gpt-4':
                            return 'GPT-4';
                        case 'gpt-3.5-turbo':
                            return 'ChatGPT';
                        case 'gpt-4o-mini':
                            return 'GPT-4o-mini';
                        default:
                            return 'OpenAI Assistant';
                    }
                default:
                    return 'Assistente';
            }
        }

        // Manipulador para a mudança de provedor LLM
        function handleProviderChange() {
            const provider = llmProviderSelect.value;
            // Armazena a preferência no localStorage
            localStorage.setItem('llmProvider', provider);
            llmProvider = provider;
            assistantName = getAssistantName(llmProvider, modelName);
        }

        // Verifica se um chat existe
        function isChatExists(chatID) {
            const chatList = JSON.parse(localStorage.getItem('chatList')) || [];
            return chatList.some(chat => chat.id === chatID);
        }

        // Manipulador para o envio do formulário de chat
        function handleFormSubmit(e) {
            e.preventDefault();
            const message = userInput.value.trim();
            if (message) {
                addMessage('Você', message, false, true);
                sendMessageToServer(message);
                userInput.value = '';
                userInput.style.height = 'auto';
            }
        }

        // Função para disparar o evento de submit de forma compatível
        function triggerSubmitEvent() {
            if (typeof Event === 'function') {
                // Compatível com navegadores modernos
                const event = new Event('submit', { cancelable: true });
                chatForm.dispatchEvent(event);
            } else {
                // Fallback para navegadores antigos
                const event = document.createEvent('Event');
                event.initEvent('submit', true, true);
                chatForm.dispatchEvent(event);
            }
        }

        // Manipulador para a tecla pressionada no campo de entrada do usuário
        function handleUserInputKeyDown(e) {
            if (e.key === 'Enter') {
                if (!e.shiftKey && !e.ctrlKey && !e.altKey && !e.metaKey) {
                    // Se Enter é pressionado sem modificadores, envie a mensagem
                    e.preventDefault(); // Impede a inserção de uma nova linha
                    triggerSubmitEvent();
                }
                // Se Shift+Enter, permite a inserção de uma nova linha
            }
        }

        // Função para ajustar a altura do textarea automaticamente
        function autoResizeTextarea() {
            this.style.height = 'auto';
            this.style.height = `${this.scrollHeight}px`;
        }

        // Debounce para limitar a frequência de execução de funções
        function debounce(fn, delay) {
            let timeoutID;
            return function (...args) {
                clearTimeout(timeoutID);
                timeoutID = setTimeout(() => fn.apply(this, args), delay);
            };
        }

        // Manipulador para o evento de scroll nas mensagens
        function handleMessagesScroll() {
            if (messagesDiv.scrollHeight - messagesDiv.scrollTop <= messagesDiv.clientHeight + 50) {
                shouldAutoScroll = true;
            } else {
                shouldAutoScroll = false;
            }
        }

        // Throttle para limitar a frequência de execução de funções
        function throttle(func, limit) {
            let lastFunc;
            let lastRan;
            return function (...args) {
                const context = this;
                if (!lastRan) {
                    func.apply(context, args);
                    lastRan = Date.now();
                } else {
                    clearTimeout(lastFunc);
                    lastFunc = setTimeout(function () {
                        if ((Date.now() - lastRan) >= limit) {
                            func.apply(context, args);
                            lastRan = Date.now();
                        }
                    }, limit - (Date.now() - lastRan));
                }
            };
        }

        // Manipulador para o botão "Nova Conversa"
        function handleNewChat() {
            currentChatID = createNewChat();
            loadChatHistory();
        }

        // Alterna a visibilidade da barra lateral
        function toggleSidebar() {
            if (sidebar.classList.contains('visible')) {
                sidebar.classList.remove('visible');
                sidebar.classList.add('hidden');
                toggleSidebarButton.innerHTML = '<i class="fas fa-bars"></i>'; // Ícone de menu
                toggleSidebarButton.setAttribute('aria-label', 'Mostrar barra lateral');
            } else {
                sidebar.classList.remove('hidden');
                sidebar.classList.add('visible');
                toggleSidebarButton.innerHTML = '<i class="fas fa-times"></i>'; // Ícone de fechar
                toggleSidebarButton.setAttribute('aria-label', 'Ocultar barra lateral');
            }
        }

        // Função para adicionar uma mensagem ao chat
        function addMessage(sender, text, isMarkdown = false, save = true, isTyping = false) {
            const messageElement = document.createElement('div');
            messageElement.classList.add('message', sender === 'Você' ? 'user-message' : 'llm-message');

            messageElement.innerHTML = `<strong>${sender}:</strong> <span class="message-content"></span>`;
            messagesDiv.appendChild(messageElement);

            const contentElement = messageElement.querySelector('.message-content');

            if (isTyping) {
                typeTextWithMarkdown(contentElement, text, isMarkdown, () => {
                    if (save) {
                        saveMessage(sender, text, isMarkdown);
                    }
                });
            } else {
                if (isMarkdown) {
                    const rawHtml = marked.parse(text);
                    const cleanHtml = DOMPurify.sanitize(rawHtml);
                    contentElement.innerHTML = cleanHtml;

                    // Destaque de sintaxe
                    elementHighlight();

                } else {
                    const cleanHtml = DOMPurify.sanitize(text);
                    contentElement.innerHTML = cleanHtml;
                }

                if (save) {
                    saveMessage(sender, text, isMarkdown);
                }
            }

            if (shouldAutoScroll) {
                messagesDiv.scrollTop = messagesDiv.scrollHeight;
            }
        }

        // Função para destacar a sintaxe após adicionar conteúdo
        function elementHighlight() {
            if (typeof hljs !== 'undefined') {
                hljs.highlightAll();
            }
        }

        // Função para exibir o texto gradualmente com formatação Markdown
        function typeTextWithMarkdown(element, fullText, isMarkdown, callback) {
            let index = 0;
            const speed = 10; // Velocidade de digitação em ms
            const increment = 2;
            let previousText = '';

            function typing() {
                if (index <= fullText.length) {
                    const newText = fullText.substring(index, index + increment);
                    previousText += newText;
                    index += increment;

                    if (isMarkdown) {
                        const tempText = closeOpenBlocks(previousText);
                        const rawHtml = marked.parse(tempText);
                        const cleanHtml = DOMPurify.sanitize(rawHtml);
                        element.innerHTML = cleanHtml;

                        // Destaque de sintaxe
                        elementHighlight();
                    } else {
                        element.textContent = previousText;
                    }

                    if (shouldAutoScroll) {
                        messagesDiv.scrollTop = messagesDiv.scrollHeight;
                    }

                    setTimeout(typing, speed);
                } else {
                    if (callback) callback();
                }
            }

            typing();
        }

        // Função para fechar blocos abertos temporariamente
        function closeOpenBlocks(text) {
            // Contar ocorrências de blocos de código
            const codeBlockMatches = text.match(/```/g);
            const codeBlockCount = codeBlockMatches ? codeBlockMatches.length : 0;

            // Se há um bloco de código não fechado, adicionar fechamento temporário
            if (codeBlockCount % 2 !== 0) {
                text += '\n```';
            }

            // Você pode adicionar lógica adicional para outros elementos de Markdown, se necessário

            return text;
        }

        // Envia a mensagem para o servidor
        async function sendMessageToServer(message) {
            try {
                const conversationHistory = getConversationHistory();

                // Adicionar indicador de carregamento
                addMessage(assistantName, 'Pensando<span class="dots">...</span>', false, false, false);

                const response = await fetch('/send', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        provider: llmProvider, // Incluímos o provedor na requisição
                        prompt: message,
                        history: conversationHistory
                    })
                });

                if (!response.ok) {
                    const errorText = await response.text();
                    throw new Error(errorText);
                }

                const data = await response.json();
                removeLastMessage();
                addMessage(assistantName, data.response, true, true, true);
            } catch (error) {
                console.error("Erro ao enviar mensagem:", error);
                removeLastMessage();
                addMessage('Erro', 'Ocorreu um erro ao enviar a mensagem. Por favor, tente novamente. ' + error, false, true);
            }
        }

        // Obtém o histórico da conversa atual
        function getConversationHistory() {
            if (!currentChatID) {
                console.error("currentChatID não está definido.");
                return [];
            }

            const history = JSON.parse(localStorage.getItem(currentChatID)) || [];
            const conversation = [];

            history.forEach(msg => {
                if (msg.sender === 'Você') {
                    conversation.push({ role: 'user', content: msg.text });
                } else if (msg.sender === assistantName) {
                    conversation.push({ role: 'assistant', content: msg.text });
                }
                // Ignorar mensagens do tipo 'Erro' ou outros remetentes
            });

            return conversation;
        }

        // Remove a última mensagem (indicador de carregamento)
        function removeLastMessage() {
            const messages = messagesDiv.getElementsByClassName('message');
            if (messages.length > 0) {
                messagesDiv.removeChild(messages[messages.length - 1]);
            }
        }

        // Salva a mensagem no localStorage
        function saveMessage(sender, text, isMarkdown) {
            if (!currentChatID) {
                console.error("currentChatID não está definido.");
                return;
            }
            const history = JSON.parse(localStorage.getItem(currentChatID)) || [];
            history.push({ sender, text, isMarkdown });
            localStorage.setItem(currentChatID, JSON.stringify(history));
        }

        // Carrega o histórico de mensagens do chat atual
        function loadChatHistory() {
            messagesDiv.innerHTML = '';
            const history = JSON.parse(localStorage.getItem(currentChatID)) || [];
            history.forEach(msg => {
                addMessage(msg.sender, msg.text, msg.isMarkdown, false);
            });
            // Atualizar o chat atual no localStorage
            localStorage.setItem('currentChatID', currentChatID);
        }

        // Limpa o histórico de mensagens do chat atual
        function clearChatHistory() {
            if (!currentChatID) return;
            localStorage.removeItem(currentChatID);
            messagesDiv.innerHTML = '';
        }

        // Carrega a lista de chats salvos
        function loadChatList() {
            chatListDiv.innerHTML = '';
            const chatList = JSON.parse(localStorage.getItem('chatList')) || [];
            chatList.forEach((chat, index) => {
                const chatItem = document.createElement('div');
                chatItem.classList.add('chat-item');

                const chatNameSpan = document.createElement('span');
                chatNameSpan.classList.add('chat-name');
                chatNameSpan.textContent = chat.name || `Conversa ${index + 1}`;

                const editButton = document.createElement('button');
                editButton.classList.add('edit-chat-name');
                editButton.innerHTML = '<i class="fas fa-edit"></i>';
                editButton.title = 'Renomear conversa';

                editButton.addEventListener('click', function (e) {
                    e.stopPropagation();
                    const newName = prompt("Digite o novo nome para a conversa:", chat.name || `Conversa ${index + 1}`);
                    if (newName) {
                        renameChat(chat.id, newName);
                    }
                });

                const deleteButton = document.createElement('button');
                deleteButton.classList.add('delete-chat');
                deleteButton.innerHTML = '<i class="fas fa-trash"></i>';
                deleteButton.title = 'Apagar conversa';

                deleteButton.addEventListener('click', function (e) {
                    e.stopPropagation();
                    const confirmDelete = confirm(`Tem certeza que deseja apagar a conversa "${chat.name || `Conversa ${index + 1}`}"?`);
                    if (confirmDelete) {
                        deleteChat(chat.id);
                    }
                });

                chatItem.appendChild(chatNameSpan);
                chatItem.appendChild(editButton);
                chatItem.appendChild(deleteButton);

                chatItem.dataset.id = chat.id;
                chatItem.addEventListener('click', function () {
                    currentChatID = chat.id;
                    loadChatHistory();
                });

                chatListDiv.appendChild(chatItem);
            });
        }

        // Renomeia um chat
        function renameChat(chatID, newName) {
            const chatList = JSON.parse(localStorage.getItem('chatList')) || [];
            const chat = chatList.find(c => c.id === chatID);
            if (chat) {
                chat.name = newName;
                localStorage.setItem('chatList', JSON.stringify(chatList));
                loadChatList();
            } else {
                console.error(`Chat com ID ${chatID} não encontrado.`);
            }
        }

        // Apaga um chat
        function deleteChat(chatID) {
            let chatList = JSON.parse(localStorage.getItem('chatList')) || [];
            chatList = chatList.filter(chat => chat.id !== chatID);
            localStorage.setItem('chatList', JSON.stringify(chatList));
            localStorage.removeItem(chatID);

            loadChatList();

            if (currentChatID === chatID) {
                if (chatList.length > 0) {
                    currentChatID = chatList[0].id;
                    loadChatHistory();
                } else {
                    currentChatID = createNewChat();
                    loadChatHistory();
                }
            }
        }

        // Cria um novo chat e atualiza a lista de chats
        function createNewChat() {
            const newChatID = generateUUID();
            const chatList = JSON.parse(localStorage.getItem('chatList')) || [];

            const chatName = `Conversa ${chatList.length + 1}`;

            chatList.push({ id: newChatID, name: chatName });
            localStorage.setItem('chatList', JSON.stringify(chatList));
            localStorage.setItem('currentChatID', newChatID);
            loadChatList();
            return newChatID;
        }

        /* Funções de Alternância de Tema */

        // Função para alternar entre os temas
        function toggleTheme() {
            const toggleThemeButton = document.getElementById('toggle-theme');
            document.body.classList.toggle('dark-mode');

            if (document.body.classList.contains('dark-mode')) {
                toggleThemeButton.innerHTML = '<i class="fas fa-sun"></i>'; // Ícone para modo Light
                toggleThemeButton.setAttribute('aria-label', 'Ativar modo Light');
                localStorage.setItem('theme', 'dark');

                // Alterar o tema do Highlight.js para escuro
                highlightStyleLink.href = "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/monokai.min.css";
            } else {
                toggleThemeButton.innerHTML = '<i class="fas fa-moon"></i>'; // Ícone para modo Dark
                toggleThemeButton.setAttribute('aria-label', 'Ativar modo Dark');
                localStorage.setItem('theme', 'light');

                // Alterar o tema do Highlight.js para claro
                highlightStyleLink.href = "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/default.min.css";
            }
        }

        // Função para carregar o tema preferido do usuário
        function loadUserTheme() {
            const savedTheme = localStorage.getItem('theme');
            const toggleThemeButton = document.getElementById('toggle-theme');
            if (savedTheme === 'dark') {
                document.body.classList.add('dark-mode');
                toggleThemeButton.innerHTML = '<i class="fas fa-sun"></i>'; // Ícone para modo Light
                toggleThemeButton.setAttribute('aria-label', 'Ativar modo Light');

                // Definir o tema do Highlight.js para escuro
                highlightStyleLink.href = "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/monokai.min.css";
            } else {
                document.body.classList.remove('dark-mode');
                toggleThemeButton.innerHTML = '<i class="fas fa-moon"></i>'; // Ícone para modo Dark
                toggleThemeButton.setAttribute('aria-label', 'Ativar modo Dark');

                // Definir o tema do Highlight.js para claro
                highlightStyleLink.href = "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/default.min.css";
            }
        }

    })();
})();

