// Verifica se o script foi carregado corretamente
console.log("script.js carregado corretamente.");

// Elementos do DOM
const chatListDiv = document.getElementById('chat-list');
const chatForm = document.getElementById('chat-form');
const userInput = document.getElementById('user-input');
const messagesDiv = document.getElementById('messages');
const clearHistoryButton = document.getElementById('clear-history-button');
const newChatButton = document.getElementById('new-chat-button');
const toggleSidebarButton = document.getElementById('toggle-sidebar');
const sidebar = document.getElementById('sidebar');
const llmProvider = document.body.dataset.llmProvider;
const modelName = document.body.dataset.modelName;
const llmProviderSelect = document.getElementById('llm-provider-select');

llmProviderSelect.value = llmProvider;

let currentChatID = null;

let assistantName;

switch (llmProvider) {
    case 'STACKSPOT':
        assistantName = 'StackSpotAI';
        break;
    case 'OPENAI':
        if (modelName === 'gpt-4') {
            assistantName = 'GPT-4';
        } else if (modelName === 'gpt-3.5-turbo') {
            assistantName = 'ChatGPT';
        } else if (modelName === 'gpt-4o-mini') {
            assistantName = 'GPT-4o-mini';
        } else {
            assistantName = 'OpenAI Assistant';
        }
        break;
    default:
        assistantName = 'Assistente';
}


// Inicializar e carregar chats
document.addEventListener('DOMContentLoaded', () => {
    console.log("DOMContentLoaded fired");

    // Inicializar Highlight.js
    if (typeof hljs !== 'undefined') {
        console.log("Highlight.js está disponível");
        hljs.highlightAll();
    } else {
        console.error("hljs não está definido. Verifique a inclusão da biblioteca.");
    }

    // Carregar a lista de chats
    loadChatList();

    // Selecionar o chat atual se existir
    const storedCurrentChat = localStorage.getItem('currentChatID');
    console.log("storedCurrentChat:", storedCurrentChat);
    if (storedCurrentChat && isChatExists(storedCurrentChat)) {
        currentChatID = storedCurrentChat;
        console.log("Chat atual encontrado:", currentChatID);
        loadChatHistory();
    } else {
        // Criar um novo chat se não houver
        currentChatID = createNewChat();
        console.log("Novo chat criado:", currentChatID);
        loadChatHistory();
    }
});

llmProviderSelect.addEventListener('change', () => {
    const provider = llmProviderSelect.value;
    fetch('/change-provider', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ provider })
    })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => { throw new Error(text) });
            }
            // Atualizar a página ou ajustar o assistantName
            location.reload();
        })
        .catch(error => {
            console.error("Erro ao alterar o provedor:", error);
        });
});


// Função para verificar se um chat existe
function isChatExists(chatID) {
    const chatList = JSON.parse(localStorage.getItem('chatList')) || [];
    const exists = chatList.some(chat => chat.id === chatID);
    console.log(`Verificando se o chat ${chatID} existe: ${exists}`);
    return exists;
}

// Evento para enviar mensagem
chatForm.addEventListener('submit', function (e) {
    e.preventDefault();
    const message = userInput.value.trim();
    console.log("Mensagem enviada:", message);
    if (message) {
        addMessage('Você', message, false, true);
        sendMessageToServer(message);
        userInput.value = '';
    }
});

// Evento para limpar histórico
clearHistoryButton.addEventListener('click', () => {
    console.log(`Limpando histórico para o chat ${currentChatID}`);
    clearChatHistory();
});

// Evento para criar nova conversa
newChatButton.addEventListener('click', () => {
    console.log("Botão 'Nova Conversa' clicado.");
    currentChatID = createNewChat();
    loadChatHistory();
});

// Evento para alternar a barra lateral
toggleSidebarButton.addEventListener('click', () => {
    console.log("Botão de alternância clicado."); // Log para verificar o evento
    sidebar.classList.toggle('hidden');
    toggleSidebarButton.textContent = sidebar.classList.contains('hidden') ? '➡' : '⬅';
    console.log(`Barra lateral ${sidebar.classList.contains('hidden') ? 'ocultada' : 'exibida'}.`);
});

// Função para adicionar mensagem ao chat
/**
 * Adiciona uma mensagem ao chat.
 * @param {string} sender - O remetente da mensagem (ex: 'Você' ou 'StackSpotAI').
 * @param {string} text - O conteúdo da mensagem.
 * @param {boolean} isMarkdown - Indica se a mensagem contém Markdown.
 * @param {boolean} save - Indica se a mensagem deve ser salva no localStorage.
 */
// Função para adicionar mensagem ao chat
// Função para adicionar mensagem ao chat com efeito de digitação e formatação em tempo real
function addMessage(sender, text, isMarkdown = false, save = true, isTyping = false) {
    console.log(`Adicionando mensagem: ${sender} - ${text}`);
    const messageElement = document.createElement('div');
    messageElement.classList.add('message', sender === 'Você' ? 'user-message' : 'llm-message');

    messageElement.innerHTML = `<strong>${sender}:</strong> <span class="message-content"></span>`;
    messagesDiv.appendChild(messageElement);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;

    const contentElement = messageElement.querySelector('.message-content');

    if (isTyping) {
        typeTextWithMarkdown(contentElement, text, isMarkdown, () => {
            // Após a digitação, garantir que o texto completo está salvo
            if (save) {
                saveMessage(sender, text, isMarkdown);
            }
        });
    } else {
        if (isMarkdown) {
            // Processa e renderiza o Markdown
            const rawHtml = marked.parse(text);
            const cleanHtml = DOMPurify.sanitize(rawHtml);
            contentElement.innerHTML = cleanHtml;

            // Destaque de sintaxe
            contentElement.querySelectorAll('pre code').forEach((block) => {
                if (hljs) {
                    hljs.highlightElement(block);
                }
            });
        } else {
            // Sanitiza e define o HTML diretamente
            const cleanHtml = DOMPurify.sanitize(text);
            contentElement.innerHTML = cleanHtml;
        }

        if (save) {
            saveMessage(sender, text, isMarkdown);
        }
    }
}


const increment = 5;

// Função para exibir o texto gradualmente com formatação Markdown
function typeTextWithMarkdown(element, fullText, isMarkdown, callback) {
    let index = 0;
    const speed = 30; // Velocidade de digitação em ms (ajuste conforme necessário)

    function typing() {
        if (index <= fullText.length) {
            const partialText = fullText.substring(0, index);
            let displayText;

            if (isMarkdown) {
                // Fechar blocos abertos temporariamente
                const tempText = closeOpenBlocks(partialText);
                const rawHtml = marked.parse(tempText);
                const cleanHtml = DOMPurify.sanitize(rawHtml);
                element.innerHTML = cleanHtml;

                // Destaque de sintaxe
                element.querySelectorAll('pre code').forEach((block) => {
                    if (hljs) {
                        hljs.highlightElement(block);
                    }
                });
            } else {
                element.textContent = partialText;
            }

            index+= increment;
            setTimeout(typing, speed);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
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

    // Verificar listas, itálicos, negritos, etc., se necessário

    return text;
}

// Função para enviar mensagem para o servidor
function sendMessageToServer(message) {
    console.log("Enviando mensagem para o servidor:", message);

    // Obter o histórico do chat atual
    const conversationHistory = getConversationHistory();

    // Adicionar indicador de carregamento
    addMessage(assistantName, 'Pensando<span class="dots">...</span>', false, false, false);

    fetch('/send', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            prompt: message,
            history: conversationHistory // Enviando o histórico da conversa
        })
    })
        .then(response => {
            if (!response.ok) {
                console.error("Erro na resposta do servidor:", response.statusText);
                return response.text().then(text => { throw new Error(text) });
            }
            return response.json();
        })
        .then(data => {
            console.log("Resposta do servidor:", data);
            removeLastMessage();
            addMessage(assistantName, data.response, true, true, true);
        })
        .catch(error => {
            console.error("Erro ao enviar mensagem:", error);
            removeLastMessage();
            addMessage('Erro', error.message, false, true);
        });
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
        // Ignorar mensagens do tipo 'Erro' ou outros remetentes
    });

    return conversation;
}


// Função para remover a última mensagem (indicador de carregamento)
function removeLastMessage() {
    console.log("Removendo a última mensagem (indicador de carregamento)");
    const messages = messagesDiv.getElementsByTagName('div');
    if (messages.length > 0) {
        messagesDiv.removeChild(messages[messages.length - 1]);
    }
}

// Função para salvar mensagem no localStorage
function saveMessage(sender, text, isMarkdown) {
    if (!currentChatID) {
        console.error("currentChatID não está definido.");
        return;
    }
    const history = JSON.parse(localStorage.getItem(currentChatID)) || [];
    history.push({ sender, text, isMarkdown });
    localStorage.setItem(currentChatID, JSON.stringify(history));
    console.log(`Mensagem salva no chat ${currentChatID}:`, { sender, text, isMarkdown });
}

// Função para carregar o histórico de mensagens do chat atual
function loadChatHistory() {
    console.log(`Carregando histórico para o chat ${currentChatID}`);
    messagesDiv.innerHTML = '';
    const history = JSON.parse(localStorage.getItem(currentChatID)) || [];
    history.forEach(msg => {
        addMessage(msg.sender, msg.text, msg.isMarkdown, false); // Não salvar novamente
    });
}

// Função para limpar o histórico de mensagens do chat atual
function clearChatHistory() {
    if (!currentChatID) return;
    localStorage.removeItem(currentChatID);
    messagesDiv.innerHTML = '';
    console.log(`Histórico do chat ${currentChatID} limpo.`);
}

// Função para carregar a lista de chats salvos
function loadChatList() {
    console.log("Carregando lista de chats.");
    chatListDiv.innerHTML = '';
    const chatList = JSON.parse(localStorage.getItem('chatList')) || [];
    chatList.forEach((chat, index) => {
        const chatItem = document.createElement('div');
        chatItem.classList.add('chat-item');

        const chatNameSpan = document.createElement('span');
        chatNameSpan.classList.add('chat-name');
        chatNameSpan.textContent = `${chat.name || `Conversa ${index + 1}`}`;

        const editButton = document.createElement('button');
        editButton.classList.add('edit-chat-name');
        editButton.textContent = '✏️';
        editButton.title = 'Renomear conversa';

        editButton.addEventListener('click', (e) => {
            e.stopPropagation(); // Evita que o clique também selecione o chat
            const newName = prompt("Digite o novo nome para a conversa:", chat.name || `Conversa ${index + 1}`);
            if (newName) {
                renameChat(chat.id, newName);
            }
        });

        const deleteButton = document.createElement('button'); // Novo botão de deletar
        deleteButton.classList.add('delete-chat');
        deleteButton.textContent = '🗑️';
        deleteButton.title = 'Apagar conversa';

        deleteButton.addEventListener('click', (e) => {
            e.stopPropagation(); // Evita que o clique também selecione o chat
            const confirmDelete = confirm(`Tem certeza que deseja apagar a conversa "${chat.name || `Conversa ${index + 1}`}"?`);
            if (confirmDelete) {
                deleteChat(chat.id);
            }
        });

        chatItem.appendChild(chatNameSpan);
        chatItem.appendChild(editButton);
        chatItem.appendChild(deleteButton); // Adiciona o botão de deletar

        chatItem.dataset.id = chat.id;
        chatItem.addEventListener('click', () => {
            console.log(`Selecionando chat ${chat.id}`);
            currentChatID = chat.id;
            localStorage.setItem('currentChatID', currentChatID);
            loadChatHistory();
        });

        chatListDiv.appendChild(chatItem);
    });
}

// Função para renomear um chat
function renameChat(chatID, newName) {
    const chatList = JSON.parse(localStorage.getItem('chatList')) || [];
    const chat = chatList.find(c => c.id === chatID);
    if (chat) {
        chat.name = newName;
        localStorage.setItem('chatList', JSON.stringify(chatList));
        loadChatList();
        console.log(`Chat ${chatID} renomeado para ${newName}`);
    } else {
        console.error(`Chat com ID ${chatID} não encontrado.`);
    }
}

// Função para apagar um chat
function deleteChat(chatID) {
    let chatList = JSON.parse(localStorage.getItem('chatList')) || [];
    chatList = chatList.filter(chat => chat.id !== chatID);
    localStorage.setItem('chatList', JSON.stringify(chatList));
    localStorage.removeItem(chatID);
    console.log(`Chat ${chatID} apagado.`);

    // Atualiza a lista de chats no DOM
    loadChatList();

    // Se o chat apagado for o atual, selecione outro chat ou crie um novo
    if (currentChatID === chatID) {
        if (chatList.length > 0) {
            currentChatID = chatList[0].id;
            localStorage.setItem('currentChatID', currentChatID);
            loadChatHistory();
        } else {
            currentChatID = createNewChat();
            loadChatHistory();
        }
    }
}

// Função para criar um novo chat e atualizar a lista de chats
function createNewChat() {
    const newChatID = crypto.randomUUID(); // Usando UUID nativo
    const chatList = JSON.parse(localStorage.getItem('chatList')) || [];

    // Nomear a conversa de forma mais amigável
    const chatName = `Conversa ${chatList.length + 1}`;

    chatList.push({ id: newChatID, name: chatName });
    localStorage.setItem('chatList', JSON.stringify(chatList));
    localStorage.setItem('currentChatID', newChatID);
    loadChatList();
    return newChatID;
}
