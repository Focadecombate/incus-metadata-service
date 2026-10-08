# Automacao do Provisionamento de Infraestrutura em Nuvem para Implantacao de Sistemas

**Autor:** TREVISO, Alex
**Ano:** 2022
**Fonte:** TCC (Eng. de Controle e Automacao), UFRGS, Porto Alegre
**Orientador:** Prof. Dr. Marcelo Gotz
**PDF:** `ufrgs_iac_provisioning_tcc.pdf`

## Resumo

TCC que implementa automacao do provisionamento de infraestrutura em nuvem AWS usando Terraform (Infrastructure as Code). A arquitetura inclui VPC, ECS com Fargate (containers serverless), ECR (registro de imagens), RDS (banco relacional), e API Gateway. Como estudo de caso, implementa um sistema de gerenciamento de ponto eletronico via API HTTP em container Docker, validado com um prototipo IoT usando ESP8266 NodeMCU com modulo NFC PN532 para leitura de identificadores RFID.

## Analise

Trabalho paralelo ao seu TCC no sentido de que ambos abordam automacao de infraestrutura em nuvem, mas em camadas diferentes. O TCC da UFRGS automatiza o provisionamento (criacao) da infraestrutura usando Terraform; o seu TCC automatiza a inicializacao (configuracao) das instancias usando cloud-init + metadata service. Sao camadas complementares: primeiro provisiona-se a infraestrutura (Terraform), depois configura-se cada instancia no primeiro boot (cloud-init via metadata service). O trabalho reforca a importancia da automacao e os desafios da configuracao manual.

## Pontos-chave

- IaC (Infrastructure as Code) surgiu em 2006 com AWS e primeiras ferramentas de gestao de infraestrutura
- Terraform: declarativo, provider-agnostic, estado gerenciado via state files
- Arquitetura AWS: VPC + subnets publicas/privadas + NAT Gateway + ALB + ECS Fargate + RDS + API Gateway
- Provisionamento completo em minutos (media de execucao documentada)
- Analise de seguranca: AWS Signature para autenticacao de API, Security Groups, subnets privadas
- Analise de escalabilidade: auto-scaling sob estresse com novos containers provisionados automaticamente
- Custo estimado: documentado para a arquitetura proposta

## Informacoes Pertinentes para o TCC

- **Complementaridade:** seu TCC opera na camada seguinte — apos o provisionamento (Terraform), a inicializacao via cloud-init usando seu metadata service configura cada instancia
- **Contexto de IaC:** definicoes e historico de IaC que podem reforcar sua introducao/fundamentacao
- **Referencia brasileira:** TCC de universidade federal brasileira na mesma area de automacao de infraestrutura
- **Contraste:** este TCC usa AWS (nuvem publica com metadata service nativo); seu TCC preenche a lacuna equivalente para Incus (nuvem privada sem metadata service)

## Sugestao BibTeX

```bibtex
@monography{treviso2022automacao,
  title={Automa{\c{c}}{\~a}o do Provisionamento de Infraestrutura em Nuvem para Implanta{\c{c}}{\~a}o de Sistemas},
  author={Treviso, Alex},
  year={2022},
  school={Universidade Federal do Rio Grande do Sul},
  address={Porto Alegre},
  type={Trabalho de Conclus{\~a}o de Curso (Eng. de Controle e Automa{\c{c}}{\~a}o)}
}
```
