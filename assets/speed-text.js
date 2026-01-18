(function() {
	// DOM elements
	const textInput = document.getElementById('text-input');
	const wpmSlider = document.getElementById('wpm-slider');
	const wpmDisplay = document.getElementById('wpm-display');
	const startBtn = document.getElementById('start-btn');
	const resetBtn = document.getElementById('reset-btn');
	const wordDisplay = document.getElementById('word-display');
	const progressEl = document.getElementById('progress');
	const timeEstimateEl = document.getElementById('time-estimate');
	const statusEl = document.getElementById('status');

	// State
	let words = [];
	let currentIndex = 0;
	let isPlaying = false;
	let intervalId = null;

	// Calculate the optimal recognition point (ORP) - approximately middle of word
	function getORP(word) {
		const len = word.length;
		if (len <= 1) return 0;
		if (len <= 3) return 1;
		// For longer words, aim for roughly the middle, slightly left of center
		return Math.floor((len - 1) / 2);
	}

	// Display a word with the focus letter highlighted using safe DOM methods
	function displayWord(word) {
		// Clear existing content
		while (wordDisplay.firstChild) {
			wordDisplay.removeChild(wordDisplay.firstChild);
		}

		// Create span elements
		const beforeSpan = document.createElement('span');
		beforeSpan.className = 'before';

		const focusSpan = document.createElement('span');
		focusSpan.className = 'focus';

		const afterSpan = document.createElement('span');
		afterSpan.className = 'after';

		if (word) {
			const orpIndex = getORP(word);
			const before = word.slice(0, orpIndex);
			const focus = word[orpIndex];
			const after = word.slice(orpIndex + 1);

			// Calculate padding to center the focus letter
			const maxPadding = 10;
			const beforePadding = maxPadding - before.length;
			const afterPadding = maxPadding - after.length;
			const paddedBefore = '\u00A0'.repeat(Math.max(0, beforePadding)) + before;
			const paddedAfter = after + '\u00A0'.repeat(Math.max(0, afterPadding));

			beforeSpan.textContent = paddedBefore;
			focusSpan.textContent = focus;
			afterSpan.textContent = paddedAfter;
		}

		wordDisplay.appendChild(beforeSpan);
		wordDisplay.appendChild(focusSpan);
		wordDisplay.appendChild(afterSpan);
	}

	// Update progress display
	function updateProgress() {
		progressEl.textContent = currentIndex + ' / ' + words.length + ' words';
	}

	// Update time estimate display
	function updateTimeEstimate() {
		const text = textInput.value.trim();
		const wordCount = text ? text.split(/\s+/).filter(function(w) { return w.length > 0; }).length : 0;
		if (wordCount === 0) {
			timeEstimateEl.textContent = '';
			return;
		}
		const wpm = parseInt(wpmSlider.value, 10);
		const totalSeconds = Math.ceil((wordCount / wpm) * 60);
		const minutes = Math.floor(totalSeconds / 60);
		const seconds = totalSeconds % 60;
		if (minutes > 0) {
			timeEstimateEl.textContent = wordCount + ' words \u2022 ' + minutes + 'm ' + seconds + 's at ' + wpm + ' WPM';
		} else {
			timeEstimateEl.textContent = wordCount + ' words \u2022 ' + seconds + 's at ' + wpm + ' WPM';
		}
	}

	// Update status display
	function updateStatus(text, className) {
		statusEl.textContent = text;
		statusEl.className = 'status ' + (className || '');
	}

	// Calculate interval from WPM
	function getInterval() {
		const wpm = parseInt(wpmSlider.value, 10);
		return Math.round(60000 / wpm);
	}

	// Show next word
	function showNextWord() {
		if (currentIndex >= words.length) {
			pause();
			updateStatus('Finished', '');
			startBtn.textContent = 'Restart';
			return;
		}

		displayWord(words[currentIndex]);
		currentIndex++;
		updateProgress();
	}

	// Start playing
	function play() {
		if (words.length === 0) return;

		isPlaying = true;
		updateStatus('Playing', 'playing');
		startBtn.textContent = 'Pause';

		// Show first word immediately
		showNextWord();

		// Set interval for subsequent words
		intervalId = setInterval(showNextWord, getInterval());
	}

	// Pause
	function pause() {
		isPlaying = false;
		updateStatus('Paused', 'paused');
		startBtn.textContent = 'Resume';

		if (intervalId) {
			clearInterval(intervalId);
			intervalId = null;
		}
	}

	// Toggle play/pause
	function togglePlayPause() {
		if (words.length === 0) {
			initializeReader();
			if (words.length === 0) return;
		}

		if (currentIndex >= words.length) {
			// Restart from beginning
			currentIndex = 0;
			play();
		} else if (isPlaying) {
			pause();
		} else {
			play();
		}
	}

	// Go to previous word
	function previousWord() {
		if (words.length === 0) return;

		// If playing, pause first
		if (isPlaying) {
			pause();
		}

		// Go back (currentIndex is already pointing to next word)
		currentIndex = Math.max(0, currentIndex - 2);
		displayWord(words[currentIndex]);
		currentIndex++;
		updateProgress();
	}

	// Go to next word
	function nextWord() {
		if (words.length === 0) return;

		// If playing, pause first
		if (isPlaying) {
			pause();
		}

		if (currentIndex < words.length) {
			displayWord(words[currentIndex]);
			currentIndex++;
			updateProgress();
		}
	}

	// Initialize reader with current text
	function initializeReader() {
		const text = textInput.value.trim();
		if (!text) {
			alert('Please enter some text first.');
			return;
		}

		// Split text into words, filtering out empty strings
		words = text.split(/\s+/).filter(function(w) { return w.length > 0; });
		currentIndex = 0;

		updateProgress();
		displayWord('');
		updateStatus('Ready', '');
		startBtn.textContent = 'Start';
	}

	// Reset everything
	function reset() {
		pause();
		words = [];
		currentIndex = 0;
		displayWord('');
		updateProgress();
		updateStatus('Ready', '');
		startBtn.textContent = 'Start';
	}

	// Event listeners
	wpmSlider.addEventListener('input', function() {
		wpmDisplay.textContent = wpmSlider.value;
		updateTimeEstimate();

		// If playing, restart interval with new speed
		if (isPlaying) {
			clearInterval(intervalId);
			intervalId = setInterval(showNextWord, getInterval());
		}
	});

	startBtn.addEventListener('click', function() {
		if (words.length === 0) {
			initializeReader();
			if (words.length > 0) {
				play();
			}
		} else {
			togglePlayPause();
		}
	});

	resetBtn.addEventListener('click', reset);

	// Keyboard controls
	document.addEventListener('keydown', function(e) {
		// Don't capture keys when typing in textarea
		if (document.activeElement === textInput) return;

		switch (e.code) {
			case 'Space':
				e.preventDefault();
				togglePlayPause();
				break;
			case 'ArrowLeft':
				e.preventDefault();
				previousWord();
				break;
			case 'ArrowRight':
				e.preventDefault();
				nextWord();
				break;
			case 'ArrowUp':
				e.preventDefault();
				wpmSlider.value = Math.min(800, parseInt(wpmSlider.value, 10) + 5);
				wpmSlider.dispatchEvent(new Event('input'));
				break;
			case 'ArrowDown':
				e.preventDefault();
				wpmSlider.value = Math.max(100, parseInt(wpmSlider.value, 10) - 5);
				wpmSlider.dispatchEvent(new Event('input'));
				break;
		}
	});

	// Update estimate when textarea changes
	textInput.addEventListener('input', updateTimeEstimate);

	// Initialize
	displayWord('');
	updateProgress();
	updateTimeEstimate();
})();
