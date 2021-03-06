package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

//Comando que genera un archivo txt a partir de una lista de archivos a buscar en el los directorios del sistema de archivos
func Cat(id string, files []string) {
	//Obtenemos la particion a partir del id
	path, mountedPart, err := searchPartition(id)
	if err != nil {
		return
	}
	//Obtenemos el file del disco
	file, _, err := readFile(path)
	if err != nil {
		return
	}
	//Definimos el tipo de particion que es
	indexSB, _ := getPartitionType(mountedPart)
	//Recuperamos el superbloque de la particion
	sb := getSB(file, indexSB)
	//Recuperamos el arbol de directorio de '/'
	root := getVirtualDirectotyTree(file, sb.PrDirectoryTree, 0)
	var data string
	var txt string
	//Recorremos todos los archivos a buscar
	for i, route := range files {
		aux := getDataFile(file, sb, root, route)
		if aux != "" {
			data += " (" + strconv.Itoa(i+1) + ") "
			data += " [" + route + "]-> "
			data += aux + "\n"
			txt += " (" + strconv.Itoa(i+1) + ") " + aux + "\n"
		}
	}
	file.Close()
	fmt.Println("[-] Estos son los datos en los archivos a buscar:")
	fmt.Print(data)
	// Escribir la data de los archivos en un txt
	err = ioutil.WriteFile("files.txt", []byte(txt), 0644)
	if err != nil {
		panic(err)
	}
}

func getDataFile(file *os.File, sb superBoot, vdt virtualDirectoryTree, route string) string {
	//Revismos que la ruta a insertar sea correcta
	if route[0] != '/' {
		fmt.Println("[ERROR] El path no es valido.")
		return ""
	}
	//Obtenemos las carpetas
	folders := strings.Split(route, "/")
	folders = folders[1:]
	//Obtenemos el nombre del archivo
	var filename [20]byte
	copy(filename[:], folders[len(folders)-1])
	folders = folders[:len(folders)-1]
	//Procedemos a obtener el puntero del DD del directorio
	var index int64
	if len(folders) > 0 {
		index = existDetailDirectory(file, &sb, vdt, folders, 0)
	} else {
		index = vdt.PrDirectoryDetail
	}
	if index == -1 {
		fmt.Println("[ERROR]: El directorio", route, "no existe.")
		return ""
	}
	//Obtenemos el detalle de directorio
	dd := getDirectotyDetail(file, sb.PrDirectoryDetail, index)
	//Recuperamos el puntero del inodo donde se encuentra el archivo
	nInode, _ := searchFile(file, sb, dd, filename)
	if nInode == -1 {
		fmt.Println("[ERROR]: El archivo", route, "no existe.")
		return ""
	}
	//Obtenemos el inodo
	inode := getInode(file, sb.PrInodeTable, nInode)
	//Recuperamos los bloques de datos
	return searchDataFile(file, sb, inode)
}

//Funcion que recorre todo el inodo y va recuperando la data de cada bloque asignado
func searchDataFile(file *os.File, sb superBoot, inode iNode) string {
	data := ""
	for i := 0; i < len(inode.Blocks); i++ {
		if inode.Blocks[i] != -1 {
			block := getBlock(file, sb.PrBlocks, inode.Blocks[i])
			data += strings.Replace(string(block.Data[:]), "\x00", "", -1)
		}
	}
	if inode.PrIndirect != -1 {
		aux := getInode(file, sb.PrDirectoryDetail, inode.PrIndirect)
		data += searchDataFile(file, sb, aux)
	}
	return data
}
