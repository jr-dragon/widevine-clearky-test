// Initialize Shaka Player
shaka.polyfill.installAll();

const videoEl = document.getElementById('player');
const placeholder = document.getElementById('placeholder');
const videoList = document.getElementById('videoList');
const uploadArea = document.getElementById('uploadArea');
const fileInput = document.getElementById('fileInput');
const uploadProgress = document.getElementById('uploadProgress');
const progressFill = document.getElementById('progressFill');
const progressText = document.getElementById('progressText');

let player = null;

async function initPlayer() {
  if (player) {
    await player.destroy();
  }
  player = new shaka.Player();
  await player.attach(videoEl);

  player.configure({
    drm: {
      servers: {
        'org.w3.clearkey': '/api/license'
      }
    }
  });

  player.addEventListener('error', (e) => {
    console.error('Player error:', e.detail);
  });
}

async function playVideo(manifestUrl) {
  placeholder.style.display = 'none';
  videoEl.style.display = 'block';

  await initPlayer();

  try {
    await player.load(manifestUrl);
    videoEl.play();
  } catch (e) {
    console.error('Failed to load manifest:', e);
    alert('播放失敗: ' + e.message);
  }
}

async function loadVideoList() {
  try {
    const resp = await fetch('/api/videos');
    const videos = await resp.json();

    if (videos.length === 0) {
      videoList.innerHTML = '<div class="empty-state">尚無加密影片，請上傳一個影片開始</div>';
      return;
    }

    videoList.innerHTML = videos.map(v => `
      <div class="video-item" data-manifest="${v.manifest}">
        <span class="name">${escapeHtml(v.name)}</span>
        <button class="play-btn">播放</button>
      </div>
    `).join('');

    videoList.querySelectorAll('.video-item').forEach(item => {
      item.addEventListener('click', () => {
        playVideo(item.dataset.manifest);
      });
    });
  } catch (e) {
    console.error('Failed to load video list:', e);
  }
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

// Upload handling
uploadArea.addEventListener('click', () => fileInput.click());
uploadArea.addEventListener('dragover', (e) => {
  e.preventDefault();
  uploadArea.classList.add('dragover');
});
uploadArea.addEventListener('dragleave', () => {
  uploadArea.classList.remove('dragover');
});
uploadArea.addEventListener('drop', (e) => {
  e.preventDefault();
  uploadArea.classList.remove('dragover');
  if (e.dataTransfer.files.length > 0) {
    uploadFile(e.dataTransfer.files[0]);
  }
});
fileInput.addEventListener('change', () => {
  if (fileInput.files.length > 0) {
    uploadFile(fileInput.files[0]);
  }
});

async function uploadFile(file) {
  if (!file.name.toLowerCase().endsWith('.mp4')) {
    alert('請上傳 MP4 格式的影片');
    return;
  }

  const formData = new FormData();
  formData.append('video', file);

  uploadProgress.style.display = 'block';
  progressFill.style.width = '0%';
  progressText.textContent = '上傳與加密中...';

  try {
    const xhr = new XMLHttpRequest();

    xhr.upload.addEventListener('progress', (e) => {
      if (e.lengthComputable) {
        const pct = Math.round((e.loaded / e.total) * 80);
        progressFill.style.width = pct + '%';
        if (pct >= 80) {
          progressText.textContent = '加密打包中，請稍候...';
        }
      }
    });

    const result = await new Promise((resolve, reject) => {
      xhr.onload = () => {
        if (xhr.status === 200) {
          resolve(JSON.parse(xhr.responseText));
        } else {
          reject(new Error(xhr.responseText));
        }
      };
      xhr.onerror = () => reject(new Error('Upload failed'));
      xhr.open('POST', '/api/encrypt');
      xhr.send(formData);
    });

    progressFill.style.width = '100%';
    progressText.textContent = '完成！正在載入影片列表...';

    await loadVideoList();

    setTimeout(() => {
      uploadProgress.style.display = 'none';
      fileInput.value = '';
    }, 1500);

    // Auto-play the newly encrypted video
    playVideo(result.manifest);

  } catch (e) {
    progressText.textContent = '失敗: ' + e.message;
    progressFill.style.width = '0%';
    console.error('Upload failed:', e);
  }
}

// Initial setup
videoEl.style.display = 'none';

if (shaka.Player.isBrowserSupported()) {
  loadVideoList();
} else {
  placeholder.textContent = '您的瀏覽器不支援 EME，請使用 Chrome 或 Firefox';
}
