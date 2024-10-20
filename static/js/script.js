document.addEventListener('DOMContentLoaded', () => {
    // Variáveis do DOM
    const chatListDiv = document.getElementById('chat-list');
    const chatForm = document.getElementById('chat-form');
    const userInput = document.getElementById('user-input');
    const messagesDiv = document.getElementById('messages');
    const newChatButton = document.getElementById('new-chat-button');
    const toggleSidebarButton = document.getElementById('toggle-sidebar');
    const sidebar = document.getElementById('sidebar');
    const llmProviderSelect = document.getElementById('llm-provider-select');
    const toggleThemeButton = document.getElementById('toggle-theme');
    const highlightStyleLink = document.getElementById('highlight-style');
    const clearHistoryButton = document.getElementById('clear-history-button');
    const chatContainer = document.getElementById('chat-container');
    const toggleSidebarButtonHidden = document.getElementById('toggle-sidebar-hidden');
    const toggleThemeButtonHidden = document.getElementById('toggle-theme-hidden');

    // Estado do aplicativo
    let currentChatID = null;
    let llmProvider = localStorage.getItem('llmProvider') || 'STACKSPOT';
    let modelName = document.body.getAttribute('data-model-name') || 'gpt-4o-mini';
    let assistantName = getAssistantName(llmProvider, modelName);
    let shouldAutoScroll = true; // Controla se o scroll automático está ativo

    // eventos
    toggleSidebarButtonHidden.addEventListener('click', toggleSidebar);
    toggleThemeButtonHidden.addEventListener('click', toggleTheme);

    // Função para gerar UUID
// Função para gerar um UUID
    function generateUUID() {
        let d = new Date().getTime();
        if (typeof performance !== 'undefined' && typeof performance.now === 'function') {
            d += performance.now(); // Usa o timer de alta precisão se disponível
        }
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
            const r = (d + Math.random() * 16) % 16 | 0;
            d = Math.floor(d / 16);
            return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
        });
    }

// Gerar um session_id se não existir
    let sessionId = localStorage.getItem('session_id');
    if (!sessionId) {
        sessionId = generateUUID();
        localStorage.setItem('session_id', sessionId);
    }


    // Inicialização
    initialize();

    function initialize() {
        // Configurar o seletor de provedor LLM
        llmProviderSelect.value = llmProvider;

        // Atualizar o assistantName
        assistantName = getAssistantName(llmProvider, modelName);

        // Carregar o tema do usuário
        loadUserTheme();

        // Carregar a lista de chats
        loadChatList();

        // Ajustar o contêiner do chat com base no estado inicial da barra lateral
        if (sidebar.classList.contains('hidden')) {
            chatContainer.classList.add('full-width');
            toggleSidebarButton.innerHTML = '<i class="fas fa-bars"></i>';
            toggleSidebarButton.setAttribute('aria-label', 'Mostrar barra lateral');
        } else {
            chatContainer.classList.remove('full-width');
            toggleSidebarButton.innerHTML = '<i class="fas fa-times"></i>';
            toggleSidebarButton.setAttribute('aria-label', 'Ocultar barra lateral');
        }

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

    function addEventListeners() {
        llmProviderSelect.addEventListener('change', handleProviderChange);
        chatForm.addEventListener('submit', handleFormSubmit);
        userInput.addEventListener('keydown', handleUserInputKeyDown);
        userInput.addEventListener('input', debounce(autoResizeTextarea, 50));
        messagesDiv.addEventListener('scroll', throttle(handleMessagesScroll, 100));
        newChatButton.addEventListener('click', handleNewChat);
        toggleSidebarButton.addEventListener('click', toggleSidebar);
        toggleThemeButton.addEventListener('click', toggleTheme);
        clearHistoryButton.addEventListener('click', clearChatHistory);
        // Adiciona o listener para detectar quando o usuário faz scroll manualmente
        messagesDiv.addEventListener('scroll', () => {
            checkIfShouldAutoScroll();
        });
    }

    function getAssistantName(provider, model) {
        switch (provider) {
            case 'STACKSPOT':
                return 'StackSpotAI';
            case 'OPENAI':
                if (model === 'gpt-4') {
                    return 'GPT-4';
                } else if (model === 'gpt-3.5-turbo') {
                    return 'ChatGPT';
                } else if (model === 'gpt-4o-mini') {
                    return 'GPT-4o-mini';
                } else {
                    return 'OpenAI Assistant';
                }
            default:
                return 'Assistente';
        }
    }

    // Função para verificar se o usuário está perto do final do chat
    function checkIfShouldAutoScroll() {
        const threshold = 50; // Distância do final para ativar o autoscroll
        const position = messagesDiv.scrollTop + messagesDiv.clientHeight;
        const height = messagesDiv.scrollHeight;
        shouldAutoScroll = height - position < threshold;  // Ativar autoscroll somente se o usuário estiver perto do final
    }

    function handleProviderChange() {
        llmProvider = llmProviderSelect.value;
        localStorage.setItem('llmProvider', llmProvider);
        assistantName = getAssistantName(llmProvider, modelName);
    }

    function isChatExists(chatID) {
        const chatList = JSON.parse(localStorage.getItem('chatList')) || [];
        return chatList.some(chat => chat.id === chatID);
    }

    function handleFormSubmit(e) {
        e.preventDefault();
        const message = userInput.value.trim();
        if (message) {
            addMessage('Você', message, 'user-message', false, true);
            sendMessageToServer(message);
            userInput.value = '';
            userInput.style.height = 'auto';
        }
    }

    function triggerSubmitEvent() {
        if (typeof Event === 'function') {
            const event = new Event('submit', { cancelable: true });
            chatForm.dispatchEvent(event);
        } else {
            const event = document.createEvent('Event');
            event.initEvent('submit', true, true);
            chatForm.dispatchEvent(event);
        }
    }

    function handleUserInputKeyDown(e) {
        if (e.key === 'Enter') {
            if (!e.shiftKey && !e.ctrlKey && !e.altKey && !e.metaKey) {
                e.preventDefault();
                triggerSubmitEvent();
            }
        }
    }

    function autoResizeTextarea() {
        this.style.height = 'auto';
        this.style.height = `${this.scrollHeight}px`;
    }

    function debounce(fn, delay) {
        let timeoutID;
        return function (...args) {
            clearTimeout(timeoutID);
            timeoutID = setTimeout(() => fn.apply(this, args), delay);
        };
    }

    function handleMessagesScroll() {
        if (messagesDiv.scrollHeight - messagesDiv.scrollTop <= messagesDiv.clientHeight + 50) {
            shouldAutoScroll = true;
        } else {
            shouldAutoScroll = false;
        }
    }

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

    function handleNewChat() {
        currentChatID = createNewChat();
        loadChatHistory();
    }

    function toggleSidebar() {
        if (sidebar.classList.contains('hidden')) {
            sidebar.classList.remove('hidden');
            chatContainer.classList.remove('full-width');
        } else {
            sidebar.classList.add('hidden');
            chatContainer.classList.add('full-width');
        }
    }

    function addMessage(sender, text, messageClass, isMarkdown = false, save = true, isTyping = false) {
        const messageElement = document.createElement('div');
        messageElement.classList.add('message', messageClass);

        const contentElement = document.createElement('div');
        contentElement.classList.add('message-content');

        if (isTyping) {
            contentElement.innerHTML = `<strong>${sender}:</strong> <span class="typing-indicator"><span class="dot"></span><span class="dot"></span><span class="dot"></span></span>`;
        } else {
            if (isMarkdown) {
                const rawHtml = marked.parse(text);
                const cleanHtml = DOMPurify.sanitize(rawHtml);
                contentElement.innerHTML = `<strong>${sender}:</strong> ${cleanHtml}`;

                // Aplicar syntax highlighting se for necessário
                hljs.highlightAll();
            } else {
                const cleanHtml = DOMPurify.sanitize(text);
                contentElement.innerHTML = `<strong>${sender}:</strong> ${cleanHtml}`;
            }
        }

        messageElement.appendChild(contentElement);
        messagesDiv.appendChild(messageElement);

        if (shouldAutoScroll) {
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
        }

        if (save) {
            saveMessage(sender, text, isMarkdown);  // Salvar a mensagem no localStorage (tanto para o usuário quanto para a assistente)
        }
    }

    function elementHighlight() {
        if (typeof hljs !== 'undefined') {
            hljs.highlightAll();
        }
    }

    async function sendMessageToServer(message) {
        try {
            const conversationHistory = getConversationHistory();

            // Adicionar indicador de digitação
            addMessage(assistantName, '', 'assistant-message', false, false, true);

            const response = await fetch('/send', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    provider: llmProvider,
                    model: modelName,
                    prompt: message,
                    history: conversationHistory,
                    session_id: sessionId  // Adicionar o session_id aqui
                })
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText);
            }

            const data = await response.json();
            const messageID = data.message_id;

            // Iniciar o polling para obter a resposta
            pollForResponse(messageID);
        } catch (error) {
            console.error("Erro ao enviar mensagem:", error);
            removeLastMessage();
            addMessage('Erro', 'Ocorreu um erro ao enviar a mensagem. Por favor, tente novamente. ' + error, 'assistant-message', false, true);
        }
    }

    async function pollForResponse(messageID) {
        try {
            const response = await fetch(`/get-response?message_id=${messageID}`);
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText);
            }

            const data = await response.json();

            if (data.status === 'completed') {
                removeLastMessage(); // Remover o indicador de "pensando"

                // Criar o contêiner da mensagem da assistente
                const assistantMessageElement = document.createElement('div');
                assistantMessageElement.classList.add('message', 'assistant-message'); // Adicionar a classe da assistente

                // Criar o conteúdo da mensagem com o nome da assistente
                const contentElement = document.createElement('div');
                contentElement.classList.add('message-content');
                contentElement.innerHTML = `<strong>${assistantName}:</strong> `; // Nome da assistente já inserido

                assistantMessageElement.appendChild(contentElement);
                messagesDiv.appendChild(assistantMessageElement);

                // Iniciar a transcrição da resposta da LLM com formatação
                transcribeText(contentElement, data.response, 50);  // Transcrever o texto com o efeito de digitação, aplicando na "contentElement"

                // Salvar a mensagem da IA no localStorage
                saveMessage(assistantName, data.response, true);  // Salva a mensagem da IA
            } else if (data.status === 'processing') {
                setTimeout(() => {
                    pollForResponse(messageID);
                }, 1000);
            } else if (data.status === 'error') {
                removeLastMessage();
                addMessage('Erro', data.message, 'assistant-message', false, true);
            }
        } catch (error) {
            console.error("Erro ao obter a resposta:", error);
            removeLastMessage();
            addMessage('Erro', 'Ocorreu um erro ao obter a resposta. Por favor, tente novamente. ' + error, 'assistant-message', false, true);
        }
    }

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
        });

        return conversation;
    }

    function removeLastMessage() {
        const messages = messagesDiv.getElementsByClassName('message');
        if (messages.length > 0) {
            messagesDiv.removeChild(messages[messages.length - 1]);
        }
    }

    function saveMessage(sender, text, isMarkdown) {
        if (!currentChatID) {
            console.error("currentChatID não está definido.");
            return;
        }

        // Carregar o histórico atual do localStorage
        const history = JSON.parse(localStorage.getItem(currentChatID)) || [];

        // Adicionar a nova mensagem ao histórico
        history.push({ sender, text, isMarkdown });

        // Salvar o histórico atualizado de volta no localStorage
        localStorage.setItem(currentChatID, JSON.stringify(history));
    }

    function loadChatHistory() {
        messagesDiv.innerHTML = ''; // Limpar o container de mensagens

        // Carregar o histórico do chat atual do localStorage
        const history = JSON.parse(localStorage.getItem(currentChatID)) || [];

        // Reexibir cada mensagem do histórico
        history.forEach(msg => {
            const messageClass = msg.sender === 'Você' ? 'user-message' : 'assistant-message';
            addMessage(msg.sender, msg.text, messageClass, msg.isMarkdown, false); // Reexibir a mensagem sem salvar novamente
        });

        // Salvar o chat atual no localStorage
        localStorage.setItem('currentChatID', currentChatID);

        // Aplicar o highlight em mensagens de código
        hljs.highlightAll();
    }


    function clearChatHistory() {
        if (!currentChatID) return;
        localStorage.removeItem(currentChatID);
        messagesDiv.innerHTML = '';
    }

    function loadChatList() {
        chatListDiv.innerHTML = '';
        const chatList = JSON.parse(localStorage.getItem('chatList')) || [];
        chatList.forEach((chat, index) => {
            const chatItem = document.createElement('div');
            chatItem.classList.add('chat-item');

            const chatNameSpan = document.createElement('span');
            chatNameSpan.classList.add('chat-name');
            chatNameSpan.textContent = chat.name || `Conversa ${index + 1}`;

            const chatActionsDiv = document.createElement('div');
            chatActionsDiv.classList.add('chat-actions');

            const editButton = document.createElement('button');
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
            deleteButton.innerHTML = '<i class="fas fa-trash"></i>';
            deleteButton.title = 'Apagar conversa';

            deleteButton.addEventListener('click', function (e) {
                e.stopPropagation();
                const confirmDelete = confirm(`Tem certeza que deseja apagar a conversa "${chat.name || `Conversa ${index + 1}`}"?`);
                if (confirmDelete) {
                    deleteChat(chat.id);
                }
            });

            chatActionsDiv.appendChild(editButton);
            chatActionsDiv.appendChild(deleteButton);

            chatItem.appendChild(chatNameSpan);
            chatItem.appendChild(chatActionsDiv);

            chatItem.dataset.id = chat.id;
            chatItem.addEventListener('click', function () {
                currentChatID = chat.id;
                loadChatHistory();
            });

            chatListDiv.appendChild(chatItem);
        });
    }

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

    function transcribeText(element, text, delay = 2, charsPerTick = 10) {
        let index = 0;
        let currentText = '';

        function typeCharacter() {
            if (index < text.length) {
                // Adicionar múltiplos caracteres por vez
                currentText += text.slice(index, index + charsPerTick);
                const sanitizedHTML = DOMPurify.sanitize(marked.parse(currentText));
                element.innerHTML = `<strong>${assistantName}:</strong> ${sanitizedHTML}`;
                index += charsPerTick;

                // Somente fazer o scroll se o autoscroll estiver ativo
                if (shouldAutoScroll) {
                    messagesDiv.scrollTo({
                        top: messagesDiv.scrollHeight,
                        behavior: 'smooth',  // Faz o scroll suave
                    });
                }

                setTimeout(typeCharacter, delay);  // Delay ajustado
            } else {
                hljs.highlightAll();  // Aplicar highlight quando o texto estiver completo
            }
        }

        typeCharacter();
    }


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

    function toggleTheme() {
        document.body.classList.toggle('dark-mode');

        highlightStyleLink.onload = function() {
            hljs.highlightAll();
        };

        const isDarkMode = document.body.classList.contains('dark-mode');

        if (isDarkMode) {
            toggleThemeButton.innerHTML = '<i class="fas fa-sun"></i>';
            toggleThemeButton.setAttribute('aria-label', 'Ativar modo Light');
            toggleThemeButtonHidden.innerHTML = '<i class="fas fa-sun"></i>';  // Atualiza o ícone fora da barra lateral
            toggleThemeButtonHidden.setAttribute('aria-label', 'Ativar modo Light');
            localStorage.setItem('theme', 'dark');
            highlightStyleLink.href = "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/monokai.min.css";
        } else {
            toggleThemeButton.innerHTML = '<i class="fas fa-moon"></i>';
            toggleThemeButton.setAttribute('aria-label', 'Ativar modo Dark');
            toggleThemeButtonHidden.innerHTML = '<i class="fas fa-moon"></i>';  // Atualiza o ícone fora da barra lateral
            toggleThemeButtonHidden.setAttribute('aria-label', 'Ativar modo Dark');
            localStorage.setItem('theme', 'light');
            highlightStyleLink.href = "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/monokai.min.css";
        }
    }

    function loadUserTheme() {
        const savedTheme = localStorage.getItem('theme');
        const isDarkMode = savedTheme === 'dark';

        highlightStyleLink.onload = function() {
            hljs.highlightAll();
        };

        if (isDarkMode) {
            document.body.classList.add('dark-mode');
            toggleThemeButton.innerHTML = '<i class="fas fa-sun"></i>';
            toggleThemeButton.setAttribute('aria-label', 'Ativar modo Light');
            toggleThemeButtonHidden.innerHTML = '<i class="fas fa-sun"></i>';  // Também atualizar fora da barra lateral
            toggleThemeButtonHidden.setAttribute('aria-label', 'Ativar modo Light');
            highlightStyleLink.href = "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/monokai.min.css";
        } else {
            document.body.classList.remove('dark-mode');
            toggleThemeButton.innerHTML = '<i class="fas fa-moon"></i>';
            toggleThemeButton.setAttribute('aria-label', 'Ativar modo Dark');
            toggleThemeButtonHidden.innerHTML = '<i class="fas fa-moon"></i>';  // Também atualizar fora da barra lateral
            toggleThemeButtonHidden.setAttribute('aria-label', 'Ativar modo Dark');
            highlightStyleLink.href = "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/monokai.min.css";
        }
    }
});
