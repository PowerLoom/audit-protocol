<script>
	import axios from 'axios';
	import { onMount, onDestroy, tick } from 'svelte';
	let api_url = "http://localhost:9200";
	export const mainAPI = axios.create({
		baseURL: api_url,
		timeout: 5000
	});
	let apiKey = "";
	mainAPI.interceptors.request.use(async function (config) {
		if (apiKey){
			config.headers = {'Auth-Token': apiKey};
		}
		return config;
	}, function (error) {
		//console.error('request error?', error);
		return Promise.reject(error);
	});
	let payloads = [];
	let showModal = false;
	let requesting = false;
	let submitting = false;
	let downloadLink = "";
	let payload = "";
	const requestDownload = async(id) => {
		console.log('requesting', id);
		downloadLink = "";
		requesting = true;
		const request = await mainAPI.get('/payload/'+id);
		console.log('status', request.data.requestStatus);
		if (request.data.requestStatus == "Completed"){
			const checkRequest = await mainAPI.get('/requests/'+request.data.requestId);
			downloadLink = api_url+checkRequest.data.downloadFile;
			setTimeout(() => {showModal = true; requesting = false;}, 1000);
		} else {
			setTimeout(() => {showModal = true; requesting = false;}, 1000);
		}
	}
	const submitPayload = async() => {
		submitting = true;
		const request = await mainAPI.post('/', {payload: payload});
		console.log(request.data);
		if (request.data.recordCid){
			payload = "";
			getPayloads();
		}
		submitting = false;
	};

	const create = async() => {
		const request = await mainAPI.post('/create', JSON.stringify({hotEnabled: true}));
		console.log(request.data);
		if (request.data.apiKey){
			apiKey = request.data.apiKey;
			localStorage.setItem('powerloom_apiKey', apiKey);
		}
	};

	onMount(async () => {
		apiKey = localStorage.getItem('powerloom_apiKey');
		if (apiKey){
			getPayloads();
		}
	});

	const getPayloads = async() => {
		const request = await mainAPI.get('/payloads?retrieval=false');
		let p = [];
		for (let i=0; i<request.data.payloads.length; i++){
			p.push(request.data.payloads[i]);
		}
		payloads = p;
	}
</script>

<div class="bg-gray-100">

	<div class="min-h-screen bg-gray-100">
		<nav class="bg-white shadow-sm">
			<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between h-16">
				<div class="flex">
					<div class="flex-shrink-0 flex items-center">
						<img class="block lg:hidden px-2 h-8 w-auto" src="./img/powerloom_square.png" alt="PowerLoom logo" />
						<img class="hidden lg:block px-2 h-8 w-auto" src="./img/powerloom_square.png" alt="PowerLoom logo" />
						 Audit Protocol
					</div>
					<div class="hidden sm:ml-6 space-x-8 sm:flex">


							<a href="/" class="inline-flex items-center px-1 pt-1 border-b-2 border-indigo-500 text-sm font-medium leading-5 text-gray-900 focus:outline-none focus:border-indigo-700 transition duration-150 ease-in-out">Dashboard</a>



					</div>
				</div>
				<div class="hidden sm:ml-6 sm:flex sm:items-center">
				<button class="p-1 border-2 border-transparent text-gray-400 rounded-full hover:text-gray-500 focus:outline-none focus:text-gray-500 focus:bg-gray-100 transition duration-150 ease-in-out" aria-label="Notifications">
					<svg class="h-6 w-6" stroke="currentColor" fill="none" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
					</svg>
				</button>

				</div>
			</div>
			</div>
		</nav>
		<div class="py-10">
			<main>
			{#if apiKey}
				<div class="max-w-7xl mx-auto sm:px-6 py-5 lg:px-8">
					<div class="mt-5 md:mt-0 md:col-span-2">
						<div class="shadow sm:rounded-md sm:overflow-hidden">
							<div class="px-4 py-5 bg-white sm:p-6">

								<div class="mt-6">
									<label for="about" class="block text-sm leading-5 font-medium text-gray-700">
										Payload
									</label>
									<div class="rounded-md shadow-sm">
										<textarea id="payload" bind:value={payload} rows="3" class="form-textarea mt-1 block w-full transition duration-150 ease-in-out sm:text-sm sm:leading-5" placeholder="This form supports string but the API can handle JSON and binary"></textarea>
									</div>
								</div>
							</div>
							<div class="px-4 py-4 bg-gray-50 text-right sm:px-6">
							{#if submitting}
								<span class="px-2 shadow-sm rounded-md">
									<button type="button" class="inline-flex items-center px-4 py-2 border border-transparent text-sm leading-5 font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-500 focus:outline-none focus:shadow-outline-indigo focus:border-indigo-700 active:bg-indigo-700 transition duration-150 ease-in-out cursor-not-allowed" disabled>
									<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
									</svg>
									Processing
									</button>
								</span>
								{:else}
								<span class="px-2 shadow-sm rounded-md">
								<button type="submit" on:click={() => submitPayload()} class="inline-flex justify-center py-2 px-4 border border-transparent text-sm leading-5 font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-500 focus:outline-none focus:border-indigo-700 focus:shadow-outline-indigo active:bg-indigo-700 transition duration-150 ease-in-out">
									Submit
								</button>
								</span>
							{/if}
							</div>
						</div>
					</div>
				</div>
				<div class="max-w-7xl mx-auto sm:px-6 lg:px-8">
					<!-- Replace with your content -->
					<div class="">

				<div class="bg-white shadow overflow-hidden sm:rounded-md">
				<ul>
				{#each payloads as payload, i}
					<li class={i > 0 ? "border-t border-gray-200" : ""}>
						<a href="#" class="block hover:bg-gray-50 focus:outline-none focus:bg-gray-50 transition duration-150 ease-in-out">
						<div class="px-4 py-4 sm:px-6">
						<div class="flex items-center justify-between">
							<div class="text-sm leading-5 font-medium text-indigo-600 truncate">
							{payload.recordCid}
							</div>
							<div class="ml-2 flex-shrink-0 flex">
							{#if payload.status == "Pinned"}
							<span class="px-4 py-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
								Stored
							</span>
							{#if requesting}
								<span class="px-2 shadow-sm rounded-md">
									<button type="button" class="inline-flex items-center px-4 py-2 border border-transparent text-sm leading-5 font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-500 focus:outline-none focus:shadow-outline-indigo focus:border-indigo-700 active:bg-indigo-700 transition duration-150 ease-in-out cursor-not-allowed" disabled>
									<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
									</svg>
									Processing
									</button>
								</span>
								{:else}
								<span class="px-2 shadow-sm rounded-md">
									<button type="button" on:click={() => requestDownload(payload.recordCid)} class="inline-flex items-center px-4 py-2 border border-transparent text-sm leading-5 font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-500 focus:outline-none focus:shadow-outline-indigo focus:border-indigo-700 active:bg-indigo-700 transition duration-150 ease-in-out">
									Download
									</button>
								</span>
							{/if}
							{:else if payload.status == "PendingPinning"}
							<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-100 text-gray-800">
								Pending Storage
							</span>
							{:else}
							<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-red-100 text-red-800">
								Storage Failed
							</span>
							{/if}
							</div>
						</div>
						<div class="mt-2 sm:flex sm:justify-between">
							<div class="sm:flex">
								<a href="https://explorer-mainnet.maticvigil.com/tx/{payload.txHash}" target="_blank">
								<div class="mt-2 flex items-center text-sm leading-5 text-gray-500 sm:mt-0">
									<svg class="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" x-description="Heroicon name: location-marker" viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd" />
									</svg>
									Tx Proof
								</div>
								</a>
							</div>
							<div class="mt-2 flex items-center text-sm leading-5 text-gray-500 sm:mt-0">
							<svg class="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" x-description="Heroicon name: calendar" viewBox="0 0 20 20" fill="currentColor">
								<path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd" />
							</svg>
							<span>
								<time datetime="payload.timestamp">{new Date(payload.timestamp*1000)}</time>
							</span>
							</div>
						</div>
						</div>
					</a>
					</li>
				{/each}
				</ul>
				</div>

				</div>
					<!-- /End replace -->
				</div>
				{:else}
				<div class="max-w-7xl mx-auto sm:px-6 lg:px-8">
				<button type="button" on:click={() => create()} class="inline-flex items-center px-4 py-2 border border-transparent text-sm leading-5 font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-500 focus:outline-none focus:shadow-outline-indigo focus:border-indigo-700 active:bg-indigo-700 transition duration-150 ease-in-out">
				Create Account
				</button>
				</div>
				{/if}
			</main>
		</div>
	</div>
	{#if showModal}
	<div class="fixed z-10 inset-0 overflow-y-auto">
		<div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
			<!--
				Background overlay, show/hide based on modal state.

				Entering: "ease-out duration-300"
					From: "opacity-0"
					To: "opacity-100"
				Leaving: "ease-in duration-200"
					From: "opacity-100"
					To: "opacity-0"
			-->
			<div class="fixed inset-0 transition-opacity">
				<div class="absolute inset-0 bg-gray-500 opacity-75"></div>
			</div>

			<!-- This element is to trick the browser into centering the modal contents. -->
			<span class="hidden sm:inline-block sm:align-middle sm:h-screen"></span>&#8203;
			<!--
				Modal panel, show/hide based on modal state.

				Entering: "ease-out duration-300"
					From: "opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
					To: "opacity-100 translate-y-0 sm:scale-100"
				Leaving: "ease-in duration-200"
					From: "opacity-100 translate-y-0 sm:scale-100"
					To: "opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
			-->
			<div class="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-sm sm:w-full sm:p-6" role="dialog" aria-modal="true" aria-labelledby="modal-headline">
				<div>
					<div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100">
						<!-- Heroicon name: check -->
						<svg class="h-6 w-6 text-green-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
					</div>
					<div class="mt-3 text-center sm:mt-5">
						<h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-headline">
							Payload {#if downloadLink}Ready{:else}Request Successful{/if}
						</h3>
						<div class="mt-2">
							<p class="text-sm leading-5 text-gray-500">
								{#if downloadLink}
								<a class="text-sm leading-5 text-gray-500" href={downloadLink} target="_blank">
									Click here to download
								</a>
								{:else}
								We will email you when the payload is ready to be downloaded.
								{/if}
							</p>
						</div>
					</div>
				</div>
				<div class="mt-5 sm:mt-6">
					<span class="flex w-full rounded-md shadow-sm">
						<button type="button" on:click={() => {showModal = false;}} class="inline-flex justify-center w-full rounded-md border border-transparent px-4 py-2 bg-indigo-600 text-base leading-6 font-medium text-white shadow-sm hover:bg-indigo-500 focus:outline-none focus:border-indigo-700 focus:shadow-outline-indigo transition ease-in-out duration-150 sm:text-sm sm:leading-5">
							Go back to dashboard
						</button>
					</span>
				</div>
			</div>
		</div>
	</div>
	{/if}
</div>
