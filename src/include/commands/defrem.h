/*-------------------------------------------------------------------------
 *
 * defrem.h
 *	  POSTGRES define and remove utility definitions.
 *
 *
 * Portions Copyright (c) 1996-2019, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 * src/include/commands/defrem.h
 *
 *-------------------------------------------------------------------------
 */
#ifndef DEFREM_H
#define DEFREM_H

#include "catalog/objectaddress.h"
#include "nodes/params.h"
#include "nodes/parsenodes.h"
#include "tcop/dest.h"
#include "utils/array.h"

struct HTAB;  /* utils/hsearch.h */

/* commands/dropcmds.c */
extern void RemoveObjects(DropStmt *stmt);

/* commands/indexcmds.c */
extern ObjectAddress DefineIndex(Oid relationId,
								 IndexStmt *stmt,
								 Oid indexRelationId,
								 Oid parentIndexId,
								 Oid parentConstraintId,
								 bool is_alter_table,
								 bool check_rights,
								 bool check_not_in_use,
								 bool skip_build,
								 bool quiet,
								 bool is_new_table);
extern void ReindexIndex(ReindexStmt *stmt, bool isTopLevel);
extern Oid	ReindexTable(ReindexStmt *stmt, bool isTopLevel);
extern void ReindexMultipleTables(const char *objectName, ReindexObjectType objectKind,
								  int options, bool concurrent);
extern char *makeObjectName(const char *name1, const char *name2,
							const char *label);
extern char *ChooseRelationName(const char *name1, const char *name2,
								const char *label, Oid namespaceid,
								bool isconstraint);
extern char *ChooseRelationNameWithCache(const char *name1, const char *name2,
										 const char *label, Oid namespaceid,
										 bool isconstraint,
										 struct HTAB *cache);
extern char *ChooseIndexName(const char *tabname, Oid namespaceId,
				List *colnames, List *exclusionOpNames,
				bool primary, bool isconstraint);
extern List *ChooseIndexColumnNames(List *indexElems);
extern bool CheckIndexCompatible(Oid oldId,
								 const char *accessMethodName,
								 List *attributeList,
								 List *exclusionOpNames);
extern Oid	GetDefaultOpClass(Oid type_id, Oid am_id);
extern Oid	ResolveOpClass(List *opclass, Oid attrType,
						   const char *accessMethodName, Oid accessMethodId);

/* commands/functioncmds.c */
extern ObjectAddress CreateFunction(ParseState *pstate, CreateFunctionStmt *stmt);
extern void RemoveFunctionById(Oid funcOid);
extern void SetFunctionReturnType(Oid funcOid, Oid newRetType);
extern void SetFunctionArgType(Oid funcOid, int argIndex, Oid newArgType);
extern ObjectAddress AlterFunction(ParseState *pstate, AlterFunctionStmt *stmt);
extern ObjectAddress CreateCast(CreateCastStmt *stmt);
extern void DropCastById(Oid castOid);
extern ObjectAddress CreateTransform(CreateTransformStmt *stmt);
extern void DropTransformById(Oid transformOid);
extern void IsThereFunctionInNamespace(const char *proname, int pronargs,
									   oidvector *proargtypes, Oid nspOid);
extern void ExecuteDoStmt(DoStmt *stmt, bool atomic);
extern void ExecuteCallStmt(CallStmt *stmt, ParamListInfo params, bool atomic, DestReceiver *dest);
extern TupleDesc CallStmtResultDesc(CallStmt *stmt);
extern Oid	get_cast_oid(Oid sourcetypeid, Oid targettypeid, bool missing_ok);
extern Oid	get_transform_oid(Oid type_id, Oid lang_id, bool missing_ok);
extern void interpret_function_parameter_list(ParseState *pstate,
											  List *parameters,
											  Oid languageOid,
											  ObjectType objtype,
											  oidvector **parameterTypes,
											  ArrayType **allParameterTypes,
											  ArrayType **parameterModes,
											  ArrayType **parameterNames,
											  List **parameterDefaults,
											  Oid *variadicArgType,
											  Oid *requiredResultType);

/* commands/operatorcmds.c */
extern ObjectAddress DefineOperator(List *names, List *parameters);
extern void RemoveOperatorById(Oid operOid);
extern ObjectAddress AlterOperator(AlterOperatorStmt *stmt);

/* commands/statscmds.c */
extern ObjectAddress CreateStatistics(CreateStatsStmt *stmt);
extern void RemoveStatisticsById(Oid statsOid);
extern void UpdateStatisticsForTypeChange(Oid statsOid,
										  Oid relationOid, int attnum,
										  Oid oldColumnType, Oid newColumnType);

/* commands/aggregatecmds.c */
extern ObjectAddress DefineAggregate(ParseState *pstate, List *name, List *args, bool oldstyle,
									 List *parameters, bool replace);

/* commands/opclasscmds.c */
extern ObjectAddress DefineOpClass(CreateOpClassStmt *stmt);
extern ObjectAddress DefineOpFamily(CreateOpFamilyStmt *stmt);
extern Oid	AlterOpFamily(AlterOpFamilyStmt *stmt);
extern void RemoveOpClassById(Oid opclassOid);
extern void RemoveOpFamilyById(Oid opfamilyOid);
extern void RemoveAmOpEntryById(Oid entryOid);
extern void RemoveAmProcEntryById(Oid entryOid);
extern void IsThereOpClassInNamespace(const char *opcname, Oid opcmethod,
									  Oid opcnamespace);
extern void IsThereOpFamilyInNamespace(const char *opfname, Oid opfmethod,
									   Oid opfnamespace);
extern Oid	get_opclass_oid(Oid amID, List *opclassname, bool missing_ok);
extern Oid	get_opfamily_oid(Oid amID, List *opfamilyname, bool missing_ok);

/* commands/tsearchcmds.c */
extern ObjectAddress DefineTSParser(List *names, List *parameters);
extern void RemoveTSParserById(Oid prsId);

extern ObjectAddress DefineTSDictionary(List *names, List *parameters);
extern void RemoveTSDictionaryById(Oid dictId);
extern ObjectAddress AlterTSDictionary(AlterTSDictionaryStmt *stmt);

extern ObjectAddress DefineTSTemplate(List *names, List *parameters);
extern void RemoveTSTemplateById(Oid tmplId);

extern ObjectAddress DefineTSConfiguration(List *names, List *parameters,
										   ObjectAddress *copied);
extern void RemoveTSConfigurationById(Oid cfgId);
extern ObjectAddress AlterTSConfiguration(AlterTSConfigurationStmt *stmt);

extern text *serialize_deflist(List *deflist);
extern List *deserialize_deflist(Datum txt);

/* commands/foreigncmds.c */
extern ObjectAddress AlterForeignServerOwner(const char *name, Oid newOwnerId);
extern void AlterForeignServerOwner_oid(Oid, Oid newOwnerId);
extern ObjectAddress AlterForeignDataWrapperOwner(const char *name, Oid newOwnerId);
extern void AlterForeignDataWrapperOwner_oid(Oid fwdId, Oid newOwnerId);
extern ObjectAddress CreateForeignDataWrapper(CreateFdwStmt *stmt);
extern ObjectAddress AlterForeignDataWrapper(AlterFdwStmt *stmt);
extern void RemoveForeignDataWrapperById(Oid fdwId);
extern ObjectAddress CreateForeignServer(CreateForeignServerStmt *stmt);
extern ObjectAddress AlterForeignServer(AlterForeignServerStmt *stmt);
extern void RemoveForeignServerById(Oid srvId);
extern ObjectAddress CreateUserMapping(CreateUserMappingStmt *stmt);
extern ObjectAddress AlterUserMapping(AlterUserMappingStmt *stmt);
extern Oid	RemoveUserMapping(DropUserMappingStmt *stmt);
extern void RemoveUserMappingById(Oid umId);
extern void CreateForeignTable(CreateForeignTableStmt *stmt, Oid relid, bool skip_permission_check);
extern void ImportForeignSchema(ImportForeignSchemaStmt *stmt);
extern Datum transformGenericOptions(Oid catalogId,
									 Datum oldOptions,
									 List *options,
									 Oid fdwvalidator);

/* commands/amcmds.c */
extern ObjectAddress CreateAccessMethod(CreateAmStmt *stmt);
extern void RemoveAccessMethodById(Oid amOid);
extern Oid	get_index_am_oid(const char *amname, bool missing_ok);
extern Oid	get_table_am_oid(const char *amname, bool missing_ok);
extern Oid	get_am_oid(const char *amname, bool missing_ok);
extern char *get_am_name(Oid amOid);

/* support routines in commands/define.c */

extern char *defGetString(DefElem *def);
extern double defGetNumeric(DefElem *def);
extern bool defGetBoolean(DefElem *def);
extern int32 defGetInt32(DefElem *def);
extern int64 defGetInt64(DefElem *def);
extern List *defGetQualifiedName(DefElem *def);
extern TypeName *defGetTypeName(DefElem *def);
extern int	defGetTypeLength(DefElem *def);
extern List *defGetStringList(DefElem *def);

#endif							/* DEFREM_H */
