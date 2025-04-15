// Variáveis do jogo
const canvas = document.getElementById('pong');
const ctx = canvas.getContext('2d');

// Modo de jogo e dificuldade
let gameMode = "onePlayer"; // "onePlayer" ou "twoPlayers"
let difficulty = "easy"; // "easy", "medium" ou "hard"

// Pontuações
let player1Score = 0;
let player2Score = 0;
const winningScore = 3;
let gameOver = false;
let winner = null;

// Elementos da animação de vitória
let animationFrame = 0;
let confetti = [];

// Raquetes
const paddleWidth = 10;
const paddleHeight = 100;
let leftPaddleY = (canvas.height - paddleHeight) / 2;
let rightPaddleY = (canvas.height - paddleHeight) / 2;
const paddleSpeed = 10; // Aumentado de 8 para 10 para acompanhar a bola mais rápida

// Bola
let ballX = canvas.width / 2;
let ballY = canvas.height / 2;
let ballRadius = 10;
let ballSpeedX = 8; // Aumentado de 5 para 8
let ballSpeedY = 8; // Aumentado de 5 para 8

// Controles
let wPressed = false;
let sPressed = false;
let upPressed = false;
let downPressed = false;

// Event listeners para os controles
document.addEventListener('keydown', keyDownHandler);
document.addEventListener('keyup', keyUpHandler);
document.getElementById('restart-btn').addEventListener('click', restartGame);

// Event listeners para os botões de modo de jogo
document.getElementById('one-player-btn').addEventListener('click', () => setGameMode('onePlayer'));
document.getElementById('two-player-btn').addEventListener('click', () => setGameMode('twoPlayers'));

// Event listeners para os botões de dificuldade
document.getElementById('easy-btn').addEventListener('click', () => setDifficulty('easy'));
document.getElementById('medium-btn').addEventListener('click', () => setDifficulty('medium'));
document.getElementById('hard-btn').addEventListener('click', () => setDifficulty('hard'));

function keyDownHandler(e) {
    if (e.key === 'w' || e.key === 'W') {
        wPressed = true;
    } else if (e.key === 's' || e.key === 'S') {
        sPressed = true;
    } else if (e.key === 'ArrowUp') {
        upPressed = true;
    } else if (e.key === 'ArrowDown') {
        downPressed = true;
    }
}

function keyUpHandler(e) {
    if (e.key === 'w' || e.key === 'W') {
        wPressed = false;
    } else if (e.key === 's' || e.key === 'S') {
        sPressed = false;
    } else if (e.key === 'ArrowUp') {
        upPressed = false;
    } else if (e.key === 'ArrowDown') {
        downPressed = false;
    }
}

// Função para criar confetes para a animação de vitória
function createConfetti() {
    confetti = [];
    for (let i = 0; i < 100; i++) {
        confetti.push({
            x: Math.random() * canvas.width,
            y: Math.random() * canvas.height - canvas.height,
            size: Math.random() * 10 + 5,
            color: `rgb(${Math.floor(Math.random() * 255)}, ${Math.floor(Math.random() * 255)}, ${Math.floor(Math.random() * 255)})`,
            speed: Math.random() * 3 + 2
        });
    }
}

// Função para desenhar os confetes
function drawConfetti() {
    confetti.forEach(particle => {
        ctx.fillStyle = particle.color;
        ctx.beginPath();
        ctx.arc(particle.x, particle.y, particle.size, 0, Math.PI * 2);
        ctx.fill();

        // Atualizar posição
        particle.y += particle.speed;

        // Reposicionar partículas que saíram da tela
        if (particle.y > canvas.height) {
            particle.y = -particle.size;
            particle.x = Math.random() * canvas.width;
        }
    });
}

// Função para desenhar a mensagem de vitória
function drawVictoryMessage() {
    const playerName = winner === 1 ? "Jogador 1" : "Jogador 2";

    // Efeito pulsante baseado no frame da animação
    const scale = 1 + 0.1 * Math.sin(animationFrame * 0.1);

    ctx.save();
    ctx.translate(canvas.width / 2, canvas.height / 2);
    ctx.scale(scale, scale);

    ctx.font = "bold 36px Arial";
    ctx.fillStyle = winner === 1 ? "#4CAF50" : "#2196F3";
    ctx.textAlign = "center";
    ctx.fillText(`${playerName} VENCEU!`, 0, 0);

    ctx.font = "24px Arial";
    ctx.fillStyle = "#FFF";
    ctx.fillText("Clique em Reiniciar para jogar novamente", 0, 40);

    ctx.restore();

    animationFrame++;
}

// Função para atualizar a pontuação na tela
function updateScore() {
    document.getElementById('player1-score').textContent = player1Score;
    document.getElementById('player2-score').textContent = player2Score;
}

// Função para definir o modo de jogo
function setGameMode(mode) {
    gameMode = mode;

    // Atualizar UI
    document.getElementById('one-player-btn').classList.toggle('active', mode === 'onePlayer');
    document.getElementById('two-player-btn').classList.toggle('active', mode === 'twoPlayers');

    // Mostrar/esconder seletor de dificuldade
    document.getElementById('difficulty-selector').style.display = mode === 'onePlayer' ? 'flex' : 'none';

    // Atualizar texto dos controles
    document.getElementById('player2-controls').innerHTML = mode === 'onePlayer' ?
        "<strong>Jogador 2:</strong> Controlado pela IA" :
        "<strong>Jogador 2:</strong> Seta para cima e Seta para baixo";

    restartGame();
}

// Função para definir a dificuldade
function setDifficulty(level) {
    difficulty = level;

    // Atualizar UI
    document.getElementById('easy-btn').classList.toggle('active', level === 'easy');
    document.getElementById('medium-btn').classList.toggle('active', level === 'medium');
    document.getElementById('hard-btn').classList.toggle('active', level === 'hard');

    restartGame();
}

// Função para reiniciar o jogo
function restartGame() {
    player1Score = 0;
    player2Score = 0;
    gameOver = false;
    winner = null;
    confetti = [];
    // Redefinir a velocidade da bola para o valor inicial
    ballSpeedX = 8;
    ballSpeedY = 8;
    resetBall();
    updateScore();
}

// Função para resetar a bola no centro
function resetBall() {
    ballX = canvas.width / 2;
    ballY = canvas.height / 2;
    // Manter a direção, mas garantir que a magnitude da velocidade não seja alterada
    ballSpeedX = ballSpeedX > 0 ? -Math.abs(ballSpeedX) : Math.abs(ballSpeedX);
    ballSpeedY = Math.random() * 14 - 7; // Velocidade Y aleatória
}

// Função para verificar se alguém ganhou
function checkWinner() {
    if (player1Score >= winningScore) {
        gameOver = true;
        winner = 1;
        createConfetti();
    } else if (player2Score >= winningScore) {
        gameOver = true;
        winner = 2;
        createConfetti();
    }
}

// Função para desenhar as raquetes
function drawPaddles() {
    // Raquete esquerda (Jogador 1)
    ctx.fillStyle = "#4CAF50";
    ctx.fillRect(0, leftPaddleY, paddleWidth, paddleHeight);

    // Raquete direita (Jogador 2)
    ctx.fillStyle = "#2196F3";
    ctx.fillRect(canvas.width - paddleWidth, rightPaddleY, paddleWidth, paddleHeight);
}

// Função para desenhar a bola
function drawBall() {
    ctx.beginPath();
    ctx.arc(ballX, ballY, ballRadius, 0, Math.PI * 2);
    ctx.fillStyle = "#FFC107";
    ctx.fill();
    ctx.closePath();
}

// Função para desenhar o campo
function drawField() {
    // Fundo preto
    ctx.fillStyle = "#000";
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    // Linha central
    ctx.beginPath();
    ctx.setLineDash([10, 15]);
    ctx.moveTo(canvas.width / 2, 0);
    ctx.lineTo(canvas.width / 2, canvas.height);
    ctx.strokeStyle = "#FFF";
    ctx.stroke();
    ctx.setLineDash([]);
}

// Função para controlar a IA
function updateAI() {
    // Diferentes níveis de dificuldade para a IA
    let reactionSpeed;
    let errorMargin;
    let predictionAccuracy;

    switch (difficulty) {
        case "easy":
            reactionSpeed = 0.03; // Movimento mais lento
            errorMargin = 30;     // Maior margem de erro
            predictionAccuracy = 0.7; // Menor precisão na previsão
            break;
        case "medium":
            reactionSpeed = 0.06;
            errorMargin = 15;
            predictionAccuracy = 0.85;
            break;
        case "hard":
            reactionSpeed = 0.09; // Movimento mais rápido
            errorMargin = 5;      // Menor margem de erro
            predictionAccuracy = 0.95; // Alta precisão na previsão
            break;
    }

    // Prever onde a bola vai chegar
    let targetY = rightPaddleY + paddleHeight / 2;

    // Só reagir quando a bola estiver indo em direção à raquete da IA
    if (ballSpeedX > 0) {
        // Calcular onde a bola vai atingir o lado direito
        // Usando física básica para prever a trajetória
        let ballDistanceX = canvas.width - ballRadius - ballX;
        let bounces = Math.floor(Math.abs(ballDistanceX / ballSpeedX));

        // Calcular posição Y prevista
        let predictedY = ballY + (ballSpeedY * (ballDistanceX / ballSpeedX));

        // Ajustar para rebatidas nas bordas
        while (predictedY < 0 || predictedY > canvas.height) {
            if (predictedY < 0) {
                predictedY = -predictedY;
            } else if (predictedY > canvas.height) {
                predictedY = 2 * canvas.height - predictedY;
            }
        }

        // Adicionar erro baseado na dificuldade
        if (Math.random() > predictionAccuracy) {
            predictedY += (Math.random() * 2 - 1) * errorMargin;
        }

        targetY = predictedY;
    }

    // Mover a raquete em direção ao alvo com velocidade baseada na dificuldade
    const paddleCenter = rightPaddleY + paddleHeight / 2;
    const distanceToTarget = targetY - paddleCenter;

    // Movimento suave em direção ao alvo
    rightPaddleY += distanceToTarget * reactionSpeed * paddleSpeed;

    // Garantir que a raquete não saia da tela
    if (rightPaddleY < 0) {
        rightPaddleY = 0;
    } else if (rightPaddleY + paddleHeight > canvas.height) {
        rightPaddleY = canvas.height - paddleHeight;
    }
}

// Função principal de atualização do jogo
function update() {
    // Movimento da raquete do jogador 1
    if (wPressed && leftPaddleY > 0) {
        leftPaddleY -= paddleSpeed;
    }
    if (sPressed && leftPaddleY < canvas.height - paddleHeight) {
        leftPaddleY += paddleSpeed;
    }

    // Movimento da raquete do jogador 2 ou IA
    if (gameMode === "twoPlayers") {
        // Controle manual para o modo de 2 jogadores
        if (upPressed && rightPaddleY > 0) {
            rightPaddleY -= paddleSpeed;
        }
        if (downPressed && rightPaddleY < canvas.height - paddleHeight) {
            rightPaddleY += paddleSpeed;
        }
    } else {
        // IA controla a raquete direita no modo de 1 jogador
        updateAI();
    }

    if (!gameOver) {
        // Movimento da bola
        ballX += ballSpeedX;
        ballY += ballSpeedY;

        // Colisão com as bordas superior e inferior
        if (ballY - ballRadius < 0 || ballY + ballRadius > canvas.height) {
            ballSpeedY = -ballSpeedY;
        }

        // Colisão com as raquetes
        // Raquete esquerda (Jogador 1)
        if (ballX - ballRadius < paddleWidth && ballY > leftPaddleY && ballY < leftPaddleY + paddleHeight) {
            ballSpeedX = -ballSpeedX * 1.05; // Aumenta a velocidade em 5% a cada rebatida
            // Ajustar ângulo baseado em onde a bola atingiu a raquete
            let deltaY = ballY - (leftPaddleY + paddleHeight / 2);
            ballSpeedY = deltaY * 0.4; // Aumentado de 0.35 para 0.4
        }

        // Raquete direita (Jogador 2)
        if (ballX + ballRadius > canvas.width - paddleWidth && ballY > rightPaddleY && ballY < rightPaddleY + paddleHeight) {
            ballSpeedX = -ballSpeedX * 1.05; // Aumenta a velocidade em 5% a cada rebatida
            // Ajustar ângulo baseado em onde a bola atingiu a raquete
            let deltaY = ballY - (rightPaddleY + paddleHeight / 2);
            ballSpeedY = deltaY * 0.4; // Aumentado de 0.35 para 0.4
        }

        // Verificar se a bola saiu pela esquerda ou direita (ponto)
        if (ballX < 0) {
            // Ponto para o Jogador 2
            player2Score++;
            updateScore();
            checkWinner();
            resetBall();
        } else if (ballX > canvas.width) {
            // Ponto para o Jogador 1
            player1Score++;
            updateScore();
            checkWinner();
            resetBall();
        }
    }
}

// Função principal de desenho
function draw() {
    // Limpar o canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // Desenhar o campo
    drawField();

    // Desenhar as raquetes e a bola
    drawPaddles();

    if (!gameOver) {
        drawBall();
    } else {
        // Desenhar animação de vitória
        drawConfetti();
        drawVictoryMessage();
    }
}

// Loop principal do jogo
function gameLoop() {
    update();
    draw();
    requestAnimationFrame(gameLoop);
}

// Iniciar o jogo
gameLoop();