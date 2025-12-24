<script lang="ts">
	import '../app.css';
</script>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	<nav class="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between h-16">
				<div class="flex items-center">
					<h1 class="text-2xl font-bold text-gray-900 dark:text-white">
						nonraid<span class="text-blue-600 dark:text-blue-400">UI</span>
					</h1>
				</div>
				<div class="flex items-center">
					<span class="text-sm text-gray-500 dark:text-gray-400">
						Open Source Storage Management
					</span>
				</div>
			</div>
		</div>
	</nav>

	<main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		<slot />
	</main>

	<footer class="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 mt-12">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
			<p class="text-center text-sm text-gray-500 dark:text-gray-400">
				Powered by <a href="https://github.com/qvr/nonraid" target="_blank" class="text-blue-600 dark:text-blue-400 hover:underline">nonraid</a>
			</p>
		</div>
	</footer>
</div>
