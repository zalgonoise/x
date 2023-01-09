// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import { Configuration, OpenAIApi } from 'openai';
import * as vscode from 'vscode';

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {

	// Create a chat window in the navigation bar
	const chatWindow = vscode.window.createWebviewPanel(
		'openaiChat',
		'OpenAI Chat',
		vscode.ViewColumn.Two,
		{ enableScripts: true }
		);
		
		// Register a context menu command that sends the selected text to the OpenAI API
		// and displays the response in the chat window
		context.subscriptions.push(vscode.commands.registerCommand('openai.getCodeSuggestion', () => {
		// Get the selected text
		const text = vscode.window.activeTextEditor!.document.getText(vscode.window.activeTextEditor!.selection);
		
		// Read the user's OpenAI API key from the configuration
		const configuration = new Configuration({
			apiKey: vscode.workspace.getConfiguration().get('openai.apiKey'),
		});
		const openai = new OpenAIApi(configuration);
		
		// Use the OpenAI API to generate a code suggestion based on the selected text
		openai.createCompletion({
		model: "text-davinci-002-render",
		prompt: text,
		max_tokens: 10,
		n: 1,
		stop: '',
		temperature: 0.5,
		}).then((response) => {
		// Display the code suggestion in the chat window
		chatWindow.webview.html = `
			<html>
			<body>
				<p>${response.data.choices[0].text}</p>
			</body>
			</html>
		`;
		});
	}));
}

// This method is called when your extension is deactivated
export function deactivate() {}
