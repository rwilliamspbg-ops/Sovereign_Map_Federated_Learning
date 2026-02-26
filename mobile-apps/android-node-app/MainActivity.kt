package io.sovereignmap.node

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.NetworkCheck
import androidx.compose.material.icons.filled.Settings
import androidx.compose.runtime.*
import androidx.compose.runtime.livedata.observeAsState
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.NavController
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            SovereignNodeTheme {
                SovereignNodeApp()
            }
        }
    }
}

@Composable
fun SovereignNodeApp() {
    val navController = rememberNavController()
    val viewModel: NodeViewModel = viewModel()
    
    NavHost(navController = navController, startDestination = "home") {
        composable("home") {
            HomeScreen(viewModel, navController)
        }
        composable("settings") {
            SettingsScreen(viewModel, navController)
        }
    }
}

@Composable
fun HomeScreen(viewModel: NodeViewModel, navController: NavController) {
    val isConnected by viewModel.isConnected.observeAsState(false)
    val accuracy by viewModel.accuracy.observeAsState(0f)
    val loss by viewModel.loss.observeAsState(0f)
    val round by viewModel.round.observeAsState(0)
    val statusMessage by viewModel.statusMessage.observeAsState("Ready to join")
    val nodeID by viewModel.nodeID.observeAsState(1000)
    
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                brush = Brush.verticalGradient(
                    colors = listOf(
                        Color.Blue.copy(alpha = 0.1f),
                        Color.Magenta.copy(alpha = 0.1f)
                    )
                )
            )
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(20.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.SpaceBetween
        ) {
            // Header
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                modifier = Modifier.padding(top = 20.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.NetworkCheck,
                    contentDescription = "Network",
                    modifier = Modifier
                        .size(60.dp)
                        .padding(bottom = 8.dp),
                    tint = Color.Blue
                )
                
                Text(
                    text = "Sovereign Node",
                    fontSize = 28.sp,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.padding(bottom = 4.dp)
                )
                
                Text(
                    text = "Federated Learning on Mobile",
                    fontSize = 14.sp,
                    color = Color.Gray
                )
            }
            
            // Status Card
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(vertical = 20.dp),
                shape = RoundedCornerShape(12.dp),
                elevation = 4.dp
            ) {
                Column(
                    modifier = Modifier.padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    StatusRow("Node ID", "#$nodeID")
                    Divider()
                    
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text("Status", color = Color.Gray, fontSize = 14.sp)
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            modifier = Modifier.padding(start = 8.dp)
                        ) {
                            Box(
                                modifier = Modifier
                                    .size(8.dp)
                                    .background(
                                        if (isConnected) Color.Green else Color.Gray,
                                        shape = RoundedCornerShape(50)
                                    )
                            )
                            Spacer(modifier = Modifier.width(4.dp))
                            Text(
                                if (isConnected) "Connected" else "Offline",
                                fontWeight = FontWeight.SemiBold,
                                fontSize = 14.sp
                            )
                        }
                    }
                    
                    Divider()
                    
                    Text(
                        text = statusMessage,
                        fontSize = 12.sp,
                        color = Color.Blue,
                        maxLines = 2
                    )
                }
            }
            
            // Metrics
            if (isConnected) {
                Card(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 10.dp),
                    shape = RoundedCornerShape(12.dp),
                    backgroundColor = Color.Gray.copy(alpha = 0.05f),
                    elevation = 0.dp
                ) {
                    Column(
                        modifier = Modifier.padding(16.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        StatusRow("Round", "$round")
                        StatusRow("Accuracy", String.format("%.2f%%", accuracy * 100))
                        StatusRow("Loss", String.format("%.4f", loss))
                    }
                }
            }
            
            Spacer(modifier = Modifier.weight(1f))
            
            // Action Buttons
            Button(
                onClick = {
                    if (isConnected) {
                        viewModel.leaveNetwork()
                    } else {
                        viewModel.joinNetwork()
                    }
                },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                colors = ButtonDefaults.buttonColors(
                    backgroundColor = if (isConnected) Color.Red else Color.Blue
                ),
                shape = RoundedCornerShape(12.dp)
            ) {
                Text(
                    text = if (isConnected) "Leave Network" else "Join Network",
                    color = Color.White,
                    fontWeight = FontWeight.SemiBold,
                    fontSize = 16.sp
                )
            }
            
            Button(
                onClick = { navController.navigate("settings") },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(48.dp)
                    .padding(top = 8.dp),
                colors = ButtonDefaults.buttonColors(
                    backgroundColor = Color.Gray.copy(alpha = 0.2f)
                ),
                shape = RoundedCornerShape(12.dp)
            ) {
                Icon(Icons.Default.Settings, contentDescription = "Settings", tint = Color.Blue)
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "Settings",
                    color = Color.Blue,
                    fontWeight = FontWeight.SemiBold
                )
            }
            
            Spacer(modifier = Modifier.height(20.dp))
        }
    }
}

@Composable
fun SettingsScreen(viewModel: NodeViewModel, navController: NavController) {
    var serverURL by remember { mutableStateOf("api.sovereignmap.io:8080") }
    var byzantineMode by remember { mutableStateOf(false) }
    var epochs by remember { mutableStateOf(3) }
    
    val nodeID by viewModel.nodeID.observeAsState(1000)
    val isConnected by viewModel.isConnected.observeAsState(false)
    
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // Header
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = "Settings",
                fontSize = 20.sp,
                fontWeight = FontWeight.Bold
            )
            Button(
                onClick = { navController.popBackStack() },
                modifier = Modifier.height(36.dp),
                shape = RoundedCornerShape(8.dp)
            ) {
                Text("Close")
            }
        }
        
        // Network Section
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            elevation = 2.dp
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Network", fontWeight = FontWeight.Bold, fontSize = 14.sp)
                Spacer(modifier = Modifier.height(8.dp))
                
                OutlinedTextField(
                    value = serverURL,
                    onValueChange = { serverURL = it },
                    label = { Text("Server URL") },
                    modifier = Modifier.fillMaxWidth(),
                    enabled = !isConnected,
                    shape = RoundedCornerShape(8.dp)
                )
            }
        }
        
        // Training Section
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            elevation = 2.dp
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Training", fontWeight = FontWeight.Bold, fontSize = 14.sp)
                Spacer(modifier = Modifier.height(12.dp))
                
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("Epochs")
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Button(onClick = { if (epochs > 1) epochs-- }, modifier = Modifier.size(32.dp), enabled = !isConnected) { Text("-") }
                        Text("$epochs", modifier = Modifier.padding(0.dp, 0.dp, 8.dp, 0.dp))
                        Button(onClick = { if (epochs < 10) epochs++ }, modifier = Modifier.size(32.dp), enabled = !isConnected) { Text("+") }
                    }
                }
                
                Spacer(modifier = Modifier.height(12.dp))
                
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("Byzantine Mode")
                    Checkbox(
                        checked = byzantineMode,
                        onCheckedChange = { byzantineMode = it },
                        enabled = !isConnected
                    )
                }
            }
        }
        
        // Node Info
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            elevation = 2.dp
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Node", fontWeight = FontWeight.Bold, fontSize = 14.sp)
                Spacer(modifier = Modifier.height(8.dp))
                StatusRow("Node ID", "#$nodeID")
            }
        }
        
        Spacer(modifier = Modifier.weight(1f))
    }
}

@Composable
fun StatusRow(label: String, value: String) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(label, color = Color.Gray, fontSize = 12.sp)
        Text(value, fontWeight = FontWeight.SemiBold, fontFamily = FontFamily.Monospace, fontSize = 12.sp)
    }
}

@Composable
fun SovereignNodeTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colors = darkColors(
            primary = Color.Blue,
            primaryVariant = Color.Blue,
            secondary = Color.Magenta,
            surface = Color.White,
            background = Color.White
        ),
        typography = Typography(
            body1 = LocalTextStyle.current
        ),
        content = content
    )
}
